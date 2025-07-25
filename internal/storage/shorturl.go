package storage

import (
	"context"
	"database/sql"
	"errors"
)

// CreateShortURL creates a new short URL entry in the database
func (p *Storage) CreateShortURL(ctx context.Context, id string, longURL string) error {
	_, err := p.db.ExecContext(ctx, "INSERT INTO short_urls (id, long_url) VALUES ($1, $2)", id, longURL)

	return err
}

// DeleteShortURL deletes a short URL entry from the database by its id
func (p *Storage) DeleteShortURL(ctx context.Context, id string) error {
	_, err := p.db.ExecContext(ctx, "DELETE FROM short_urls WHERE id = $1", id)

	return err
}

// GetLongURL retrieves the long URL associated with a given short URL id
func (p *Storage) GetLongURL(ctx context.Context, id string) (string, bool, error) {
	var longURL string
	err := p.db.QueryRowContext(ctx, "SELECT long_url FROM short_urls WHERE id = $1", id).Scan(&longURL)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return "", false, nil
		}

		return "", false, err
	}

	return longURL, true, nil
}
