package cmd

import (
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strings"
	"syscall"
	"time"

	"github.com/adrg/xdg"
	"github.com/cycloidio/sqlr"
	"github.com/gorilla/handlers"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"github.com/pikoci/registry/pkreg"
	"github.com/pikoci/registry/pkreg/db"
	"github.com/pikoci/registry/pkreg/db/migrate"
	tshttp "github.com/pikoci/registry/pkreg/transport/http"
)

var serverViper = viper.New()

var serverCmd = &cobra.Command{
	Use:   "server",
	Short: "Starts the PikoCI Registry server",
	RunE: func(cmd *cobra.Command, args []string) error {
		ctx, cancel := context.WithCancel(cmd.Context())
		defer cancel()

		cfgFile, _ := cmd.Flags().GetString("config")
		if cfgFile != "" {
			serverViper.SetConfigFile(cfgFile)
			if err := serverViper.ReadInConfig(); err != nil {
				return fmt.Errorf("error loading config file: %v", err)
			}
		}

		jwtSecret := serverViper.GetString("jwt-secret")
		if jwtSecret == "" {
			jwtSecret = "dev-secret-change-me"
		}

		logLevel := serverViper.GetString("log-level")
		logger := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: parseSlogLevel(logLevel)}))
		logger = logger.With("service", "pikoci-registry")

		dbSystem := serverViper.GetString("db-system")
		if dbSystem != db.Mem && dbSystem != db.MySQL && dbSystem != db.SQLite && dbSystem != db.PostgreSQL {
			return fmt.Errorf("invalid db-system %q, should be one of: %s, %s, %s or %s", dbSystem, db.Mem, db.MySQL, db.SQLite, db.PostgreSQL)
		}

		dbFile, err := xdg.DataFile(filepath.Join(AppName, AppName+".db"))
		if err != nil {
			return fmt.Errorf("failed to create dbFile: %v", err)
		}

		logger.Info("DB connection starting ...", "db-system", dbSystem)
		database, err := db.New(
			serverViper.GetString("db-host"),
			serverViper.GetInt("db-port"),
			serverViper.GetString("db-user"),
			serverViper.GetString("db-password"),
			db.Options{
				DBName:          serverViper.GetString("db-name"),
				MultiStatements: true,
				ClientFoundRows: true,
				System:          dbSystem,
				DBFile:          dbFile,
			},
		)
		if err != nil {
			return fmt.Errorf("failed to connect to database: %w", err)
		}
		logger.Info("DB connection started", "db-system", dbSystem)

		if serverViper.GetBool("run-migrations") {
			logger.Info("Running migrations")
			if err := migrate.Migrate(database, dbSystem); err != nil {
				return fmt.Errorf("failed to run migrations: %w", err)
			}
			logger.Info("Migrations ran")
		}

		var querier sqlr.Querier = database
		if db.IsPostgreSQL(dbSystem) {
			querier = db.NewPGQuerier(database)
		}

		ur := db.NewUserRepository(querier)
		nsr := db.NewNamespaceRepository(querier)
		rtr := db.NewRegTypeRepository(querier)
		vr := db.NewVersionRepository(querier)
		tgr := db.NewTagRepository(querier)
		tkr := db.NewTokenRepository(querier)
		omr := db.NewOrgMemberRepository(querier)
		dlr := db.NewDownloadRepository(querier)

		logger.Info("Initializing service")
		svc := pkreg.New(
			ur, nsr, rtr, vr, tgr, tkr, omr, dlr,
			[]byte(jwtSecret),
			serverViper.GetString("github-client-id"),
			serverViper.GetString("github-client-secret"),
			logger,
		)
		logger.Info("Initialized service")

		logger.Info("Initializing HTTP handlers")
		handler := tshttp.Handler(svc, []byte(jwtSecret), logger.With("component", "HTTP"))
		logger.Info("Initialized HTTP handlers")

		port := serverViper.GetInt("port")
		svr := &http.Server{
			Addr:    fmt.Sprintf(":%d", port),
			Handler: handlers.CombinedLoggingHandler(os.Stdout, handler),
		}

		errs := make(chan error, 1)
		go func() {
			logger.Info("Starting HTTP server", "port", port)
			if err := svr.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				errs <- fmt.Errorf("HTTP server error: %w", err)
			}
		}()

		quit := make(chan os.Signal, 1)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

		select {
		case sig := <-quit:
			logger.Info("Received signal, shutting down", "signal", sig)
			cancel()
			shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 10*time.Second)
			defer shutdownCancel()
			svr.Shutdown(shutdownCtx)
		case err := <-errs:
			logger.Error("Component failed", "error", err)
			cancel()
			svr.Close()
			return err
		}

		_ = ctx
		return nil
	},
}

func init() {
	serverCmd.Flags().StringP("config", "c", "", "Path to the config file")
	serverCmd.Flags().IntP("port", "p", 8080, "Port in which to start the server")
	serverCmd.Flags().String("jwt-secret", "", "Secret used to sign JWT tokens")
	serverCmd.Flags().String("db-system", "mem", "Which DB system to use (mem, sqlite, mysql, postgresql)")
	serverCmd.Flags().String("db-host", "", "Database Host")
	serverCmd.Flags().Int("db-port", 0, "Database Port")
	serverCmd.Flags().String("db-user", "", "Database User")
	serverCmd.Flags().String("db-password", "", "Database Password")
	serverCmd.Flags().String("db-name", "", "Database Name")
	serverCmd.Flags().Bool("run-migrations", true, "Flag to know if migrations should be ran")
	serverCmd.Flags().String("log-level", "info", "Sets the log level ('debug', 'info', 'warn', 'error')")
	serverCmd.Flags().String("github-client-id", "", "GitHub OAuth App Client ID")
	serverCmd.Flags().String("github-client-secret", "", "GitHub OAuth App Client Secret")
	serverCmd.Flags().String("github-redirect-url", "", "GitHub OAuth Redirect URL")

	serverViper.BindPFlags(serverCmd.Flags())
	serverViper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	serverViper.AutomaticEnv()
}

func parseSlogLevel(s string) slog.Level {
	switch s {
	case "debug":
		return slog.LevelDebug
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}
