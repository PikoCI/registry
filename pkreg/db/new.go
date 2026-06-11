package db

import (
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/VividCortex/mysqlerr"
	"github.com/go-sql-driver/mysql"
	"github.com/lib/pq"

	_ "modernc.org/sqlite"
)

const (
	Mem        = "mem"
	MySQL      = "mysql"
	SQLite     = "sqlite"
	PostgreSQL = "postgresql"
)

func IsPostgreSQL(system string) bool {
	return system == PostgreSQL
}

type Options struct {
	DBName          string
	ClientFoundRows bool
	ParseTime       bool
	MultiStatements bool
	System          string
	DBFile          string
}

func New(host string, port int, user, password string, ops Options) (*sql.DB, error) {
	switch ops.System {
	case MySQL, PostgreSQL:
		if host == "" {
			return nil, errors.New("host is a required parameter")
		} else if port == 0 {
			return nil, errors.New("port is a required parameter")
		} else if user == "" {
			return nil, errors.New("user is a required parameter")
		} else if password == "" {
			return nil, errors.New("password is a required parameter")
		}
	case Mem:
	case SQLite:
		if ops.DBFile == "" {
			return nil, fmt.Errorf("DBFile is required")
		}
	default:
		return nil, fmt.Errorf("invalid db system %q", ops.System)
	}

	var (
		db  *sql.DB
		err error
	)

	switch ops.System {
	case Mem:
		db, err = sql.Open("sqlite", "file::memory:?cache=shared&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_txlock=immediate")
	case SQLite:
		db, err = sql.Open("sqlite", ops.DBFile+"?_pragma=foreign_keys(1)&_pragma=journal_mode(WAL)&_pragma=busy_timeout(5000)&_txlock=immediate")
	case PostgreSQL:
		dsn := fmt.Sprintf(
			"host=%s port=%d user=%s password=%s dbname=%s sslmode=disable",
			host, port, user, password, ops.DBName,
		)
		db, err = sql.Open("postgres", dsn)
	case MySQL:
		dsn := fmt.Sprintf(
			"%s:%s@tcp(%s:%d)/%s?clientFoundRows=%t&parseTime=%t&multiStatements=%t",
			user, password, host, port, ops.DBName, ops.ClientFoundRows, ops.ParseTime, ops.MultiStatements,
		)
		db, err = sql.Open("mysql", dsn)
	}
	if err != nil {
		return nil, fmt.Errorf("could not connect to the database: %w", err)
	}

	if err := db.Ping(); err != nil {
		if ops.System == MySQL {
			var sqlerr *mysql.MySQLError
			if errors.As(err, &sqlerr) && sqlerr.Number == mysqlerr.ER_BAD_DB_ERROR {
				ndns := fmt.Sprintf(
					"%s:%s@tcp(%s:%d)/%s?clientFoundRows=%t&parseTime=%t&multiStatements=%t",
					user, password, host, port, "", ops.ClientFoundRows, ops.ParseTime, ops.MultiStatements,
				)
				ndb, err := sql.Open("mysql", ndns)
				if err != nil {
					return nil, fmt.Errorf("could not connect to MySQL to create database: %w", err)
				}
				defer ndb.Close()
				if err := ndb.Ping(); err != nil {
					return nil, fmt.Errorf("could not ping DB to create database: %w", err)
				}
				_, err = ndb.Exec(fmt.Sprintf("CREATE DATABASE IF NOT EXISTS %s", ops.DBName))
				if err != nil {
					return nil, fmt.Errorf("could not create DB %s: %w", ops.DBName, err)
				}
				if err := db.Ping(); err != nil {
					return nil, fmt.Errorf("could not ping DB to check database created: %w", err)
				}
			} else {
				return nil, fmt.Errorf("could not ping DB: %w", err)
			}
		} else if ops.System == PostgreSQL {
			var pqerr *pq.Error
			if errors.As(err, &pqerr) && pqerr.Code == "3D000" {
				dsn := fmt.Sprintf(
					"host=%s port=%d user=%s password=%s dbname=postgres sslmode=disable",
					host, port, user, password,
				)
				ndb, err := sql.Open("postgres", dsn)
				if err != nil {
					return nil, fmt.Errorf("could not connect to PostgreSQL to create database: %w", err)
				}
				defer ndb.Close()
				if err := ndb.Ping(); err != nil {
					return nil, fmt.Errorf("could not ping PostgreSQL to create database: %w", err)
				}
				var exists bool
				err = ndb.QueryRow("SELECT EXISTS(SELECT 1 FROM pg_database WHERE datname = $1)", ops.DBName).Scan(&exists)
				if err != nil {
					return nil, fmt.Errorf("could not check if database exists: %w", err)
				}
				if !exists {
					_, err = ndb.Exec("CREATE DATABASE " + pqQuoteIdentifier(ops.DBName))
					if err != nil {
						return nil, fmt.Errorf("could not create DB %s: %w", ops.DBName, err)
					}
				}
				if err := db.Ping(); err != nil {
					return nil, fmt.Errorf("could not ping DB to check database created: %w", err)
				}
			} else {
				return nil, fmt.Errorf("could not ping DB: %w", err)
			}
		} else if ops.System != Mem && ops.System != SQLite {
			return nil, fmt.Errorf("could not ping DB: %w", err)
		}
	}

	return db, nil
}

func pqQuoteIdentifier(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `""`) + `"`
}
