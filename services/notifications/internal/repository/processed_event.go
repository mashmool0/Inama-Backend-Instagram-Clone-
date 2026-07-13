package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProcessedEventRepository struct {
	pool *pgxpool.Pool
}

func NewProcessedEventRepository(pool *pgxpool.Pool) *ProcessedEventRepository {
	return &ProcessedEventRepository{pool: pool}
}

func (r *ProcessedEventRepository) InsertIfAbsent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO processed_events (event_id, event_type)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`, eventID, eventType)
	if err != nil {
		return false, fmt.Errorf("insert processed event: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}
