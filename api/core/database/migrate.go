package database

import (
	"context"
	"database/sql"
	"log"
	"os"

	"github.com/pressly/goose/v3"
)

// migrateDB migrates the database to the latest version using goose.
func migrateDB(db *sql.DB) error {
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

		log.Println("[", i+1, "] Migration applied:", results[i].Source.Path)
	}

	return nil
}
