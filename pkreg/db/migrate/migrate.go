package migrate

import (
	"database/sql"
	"fmt"
	"strings"

	"github.com/lopezator/migrator"
	"github.com/pikoci/registry/pkreg/db"
	"github.com/pikoci/registry/pkreg/db/migrate/migrations"
)

func Migrate(database *sql.DB, system string, opts ...migrator.Option) error {
	ms := make([]interface{}, 0, len(migrations.Migrations))
	for _, m := range migrations.Migrations {
		val := m
		ms = append(ms, &migrator.Migration{
			Name: val.Name,
			Func: func(tx *sql.Tx) error {
				s := adaptSQL(val.SQL, system)
				if _, err := tx.Exec(s); err != nil {
					return err
				}
				return nil
			},
		})
	}

	allOpts := []migrator.Option{migrator.Migrations(ms...)}
	allOpts = append(allOpts, opts...)

	m, err := migrator.New(allOpts...)
	if err != nil {
		return fmt.Errorf("error while creating the migration: %w", err)
	}

	if err := m.Migrate(database); err != nil {
		return fmt.Errorf("error while migrating: %w", err)
	}

	return nil
}

func adaptSQL(sql string, system string) string {
	switch system {
	case db.Mem, db.SQLite:
		sql = strings.ReplaceAll(sql, "VARCHAR(36)", "TEXT")
		sql = strings.ReplaceAll(sql, "VARCHAR(255)", "TEXT")
		sql = strings.ReplaceAll(sql, "VARCHAR(50)", "TEXT")
		sql = strings.ReplaceAll(sql, "VARCHAR(20)", "TEXT")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT FALSE", "INTEGER DEFAULT 0")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT TRUE", "INTEGER DEFAULT 1")
		sql = strings.ReplaceAll(sql, "BOOLEAN", "INTEGER")
		sql = strings.ReplaceAll(sql, "TIMESTAMP NULL", "TEXT")
		sql = strings.ReplaceAll(sql, "TIMESTAMP", "TEXT")
		sql = strings.ReplaceAll(sql, "DATE", "TEXT")
	case db.PostgreSQL:
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT FALSE", "BOOLEAN DEFAULT FALSE")
		sql = strings.ReplaceAll(sql, "BOOLEAN DEFAULT TRUE", "BOOLEAN DEFAULT TRUE")
		sql = strings.ReplaceAll(sql, "TIMESTAMP NULL", "TIMESTAMPTZ")
		sql = strings.ReplaceAll(sql, "TIMESTAMP", "TIMESTAMPTZ")
	case db.MySQL:
		// MySQL syntax is the base, no changes needed
	}
	return sql
}
