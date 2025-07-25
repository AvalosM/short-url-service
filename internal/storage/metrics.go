package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/AvalosM/meli-interview-challenge/pkg/metrics"
)

// CreateMetrics inserts multiple metric collectors into the database
func (p *Storage) CreateMetrics(ctx context.Context, collectors map[string]*metrics.Collector) error {
	if len(collectors) == 0 {
		return nil
	}

	now := time.Now()

	queryBuilder := p.builder.
		Insert("short_url_metrics").
		Columns("short_url_id", "visit_count", "unique_visit_count", "timestamp")

	for _, collector := range collectors {
		queryBuilder = queryBuilder.Values(collector.ShortURLId, collector.Visits, collector.UniqueVisits(), now)
	}

	query, args, err := queryBuilder.ToSql()
	if err != nil {
		return fmt.Errorf("building create metrics query: %w", err)
	}

	_, err = p.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("executing create metrics query: %w", err)
	}

	return nil
}

// GetMetrics retrieves the metrics for a specific short URL ID within a given time range
func (p *Storage) GetMetrics(ctx context.Context, shortURLId string, from, to time.Time) (*metrics.Metrics, bool, error) {
	query := `SELECT SUM(visit_count), SUM(unique_visit_count) 
			  FROM short_url_metrics
			  WHERE short_url_id = $1 AND timestamp BETWEEN $2 AND $3
			  GROUP BY short_url_id`

	var visits, uniqueVisits int64
	if err := p.db.QueryRowContext(ctx, query, shortURLId, from, to).Scan(&visits, &uniqueVisits); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, false, nil
		}

		return nil, false, fmt.Errorf("executing get metrics query: %w", err)
	}

	return &metrics.Metrics{
		ShortURLId:   shortURLId,
		Visits:       visits,
		UniqueVisits: uniqueVisits,
		From:         from,
		To:           to,
	}, true, nil
}
