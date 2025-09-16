package pg

import (
	"astral/env"
	"fmt"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq" // Register PostgreSQL driver
)

const (
	DRIVER = "postgres"
)

type Storage struct {
	DB *sqlx.DB
}

func NewDBConnection(cfg *env.PgSql) (*Storage, error) {
	const op = "db.postgres.NewDBConnection"

	var connStr string
	if cfg.URI == "" {
		connStr = fmt.Sprintf("user=%s dbname=%s sslmode=%s password=%s host=%s",
			cfg.User, cfg.DbName, cfg.SSLMode, cfg.Password, cfg.Host)
	} else {
		connStr = cfg.URI
	}

	db, err := sqlx.Connect("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}
	return &Storage{DB: db}, nil
}

func (s *Storage) Stop() error {
	return s.DB.Close()
}
