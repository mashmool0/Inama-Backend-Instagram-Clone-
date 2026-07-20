package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type ProfileSyncRepository struct {
	pool *pgxpool.Pool
}

func NewProfileSyncRepository(pool *pgxpool.Pool) *ProfileSyncRepository {
	return &ProfileSyncRepository{pool: pool}
}

func (r *ProfileSyncRepository) InsertEventIfAbsent(ctx context.Context, tx pgx.Tx, eventID, eventType string) (bool, error) {
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

func (r *ProfileSyncRepository) CreateProfile(ctx context.Context, tx pgx.Tx, userID, username string) error {
	_, err := tx.Exec(ctx, `
INSERT INTO users (id, username)
VALUES ($1, $2)
ON CONFLICT (id) DO UPDATE SET username = EXCLUDED.username, updated_at = NOW()
`, userID, username)
	if err != nil {
		return fmt.Errorf("create synchronized profile: %w", err)
	}
	return nil
}

func (r *ProfileSyncRepository) UpdateUsername(ctx context.Context, tx pgx.Tx, userID, username string) error {
	tag, err := tx.Exec(ctx, `
UPDATE users SET username = $2, updated_at = NOW() WHERE id = $1
`, userID, username)
	if err != nil {
		return fmt.Errorf("update synchronized username: %w", err)
	}
	if tag.RowsAffected() != 1 {
		return fmt.Errorf("update synchronized username: profile does not exist")
	}
	return nil
}
