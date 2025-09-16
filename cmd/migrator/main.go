package main

import (
	"errors"
	"flag"
	"fmt"
	"log"
	"strings"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	var dbURL, migrationsPath, migrationsTable string
	var down bool

	flag.StringVar(&dbURL, "db-url", "", "PostgreSQL connection URL")
	flag.StringVar(&migrationsPath, "migrations-path", "", "path to migrations")
	flag.StringVar(&migrationsTable, "migrations-table", "schema_migrations", "name of migrations table")
	flag.BoolVar(&down, "down", false, "rollback the last migration")
	flag.Parse()

	if dbURL == "" {
		log.Fatal("db-url is required")
	}
	if migrationsPath == "" {
		log.Fatal("migrations-path is required")
	}

	dbURLWithTable := dbURL
	if !strings.Contains(dbURL, "?") {
		dbURLWithTable += "?"
	} else {
		dbURLWithTable += "&"
	}
	dbURLWithTable += "x-migrations-table=" + migrationsTable

	m, err := migrate.New(
		"file://"+migrationsPath,
		dbURLWithTable,
	)
	if err != nil {
		log.Fatal(err)
	}

	if down {
		if err := m.Down(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to roll back")
				return
			}
			log.Fatal(err)
		}
		fmt.Println("Migration rolled back successfully")
	} else {
		if err := m.Up(); err != nil {
			if errors.Is(err, migrate.ErrNoChange) {
				fmt.Println("No migrations to apply")
				return
			}
			log.Fatal(err)
		}
		fmt.Println("Migrations applied successfully")
	}
}
