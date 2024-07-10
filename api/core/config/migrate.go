package config

import (
	"context"
	"database/sql"
	"fmt"
	"os"

	"github.com/pressly/goose/v3"
	"go.uber.org/zap"
)

// migrateDB migrates the database to the latest version.
func migrateDB(db *sql.DB, log *zap.Logger) error {
	provider, err := goose.NewProvider(
		goose.DialectPostgres,
		db,
		os.DirFS("db/migrations"),
	)

	if err != nil {
		return err
	}

	results, err := provider.Up(context.Background())
	if err != nil {
		return err
	}

	for i := range results {
		if results[i].Error != nil {
			return results[i].Error
		}

		log.Info(
			fmt.Sprintf("[%d] Migration applied: %s", i+1, results[i].Source.Path),
		)
	}

	return nil
}
