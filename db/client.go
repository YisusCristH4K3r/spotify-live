package db

import (
	"context"
	"database/sql"

	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/sqlite"
	_ "github.com/golang-migrate/migrate/v4/source/pkger"
	"github.com/markbates/pkger"
	_ "modernc.org/sqlite"
)

func InitDatabase(ctx context.Context, db string) (*sql.DB, error) {
	//database, err := sql.Open("sqlite", ":memory:")
	database, err := sql.Open("sqlite", db)
	if err != nil {
		return nil, err
	}

	if false {
		pkger.Include("/sql/migrations")
		driver, err := sqlite.WithInstance(database, &sqlite.Config{})
		m, migrationErr := migrate.NewWithDatabaseInstance(
			"pkger:///sql/migrations",
			"sqlite", driver)
		migrationErr = m.Up()
		if migrationErr != nil {
			m.Close()
			return nil, err
		}
	}

	return database, nil
}
