package migrations

import (
	"database/sql"
	"embed"
	"fmt"

	"github.com/pressly/goose/v3"
)

//go:embed *.sql
var migrationFiles embed.FS

// RunMigrations applies all pending up migrations embedded in the binary.
// It is safe to call on every startup — already-applied migrations are skipped.
func RunMigrations(db *sql.DB) error {
	goose.SetBaseFS(migrationFiles)

	if err := goose.SetDialect("postgres"); err != nil {
		return fmt.Errorf("migration dialect: %w", err)
	}

	if err := goose.Up(db, "."); err != nil {
		return fmt.Errorf("migration up: %w", err)
	}

	return nil
}
