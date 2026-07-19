package schema

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

var statements = []string{
	`
CREATE TABLE IF NOT EXISTS notifications (
    id UUID PRIMARY KEY,
    recipient_id UUID NOT NULL,
    actor_id UUID NOT NULL,
    type SMALLINT NOT NULL,
    post_id UUID NULL,
    is_read BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`,
	`
CREATE TABLE IF NOT EXISTS processed_events (
    event_id UUID PRIMARY KEY,
    event_type TEXT NOT NULL,
    processed_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
)
`,
	`CREATE INDEX IF NOT EXISTS idx_notifications_recipient_created_id ON notifications (recipient_id, created_at DESC, id DESC)`,
	`CREATE INDEX IF NOT EXISTS idx_notifications_recipient_read_created_id ON notifications (recipient_id, is_read, created_at DESC, id DESC)`,
}

func Bootstrap(ctx context.Context, pool *pgxpool.Pool) error {
	for _, statement := range statements {
		if _, err := pool.Exec(ctx, statement); err != nil {
			return fmt.Errorf("exec schema statement: %w", err)
		}
	}

	return nil
}
