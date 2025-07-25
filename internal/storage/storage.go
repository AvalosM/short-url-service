package storage

import (
	"database/sql"

	"github.com/Masterminds/squirrel"
	_ "github.com/jackc/pgx/v5/stdlib"
)

// Storage contains resources to interact with a database.
type Storage struct {
	db      *sql.DB
	builder squirrel.StatementBuilderType
}

// NewStorage creates a new Storage
func NewStorage(config *Config) (*Storage, error) {
	db, err := sql.Open(config.Driver, config.DataSourceName)
	if err != nil {
		return nil, err
	}

	if err := db.Ping(); err != nil {
		return nil, err
	}

	return &Storage{
		db:      db,
		builder: squirrel.StatementBuilder.PlaceholderFormat(squirrel.Dollar),
	}, nil
}

// Healthy checks if the database connection is healthy.
func (p *Storage) Healthy() bool {
	return p.db.Ping() == nil
}
