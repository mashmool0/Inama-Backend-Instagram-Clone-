package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/user/internal/pagination"
)

type FollowRepository struct {
	pool *pgxpool.Pool
}

func NewFollowRepository(pool *pgxpool.Pool) *FollowRepository {
	return &FollowRepository{pool: pool}
}

func (r *FollowRepository) CreateFollow(ctx context.Context, tx pgx.Tx, followerID, followeeID string) (bool, error) {
	tag, err := tx.Exec(ctx, `
INSERT INTO follows (follower_id, followee_id)
VALUES ($1, $2)
ON CONFLICT DO NOTHING
`, followerID, followeeID)
	if err != nil {
		return false, fmt.Errorf("create follow edge: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *FollowRepository) DeleteFollow(ctx context.Context, tx pgx.Tx, followerID, followeeID string) (bool, error) {
	tag, err := tx.Exec(ctx, `
DELETE FROM follows
WHERE follower_id = $1 AND followee_id = $2
`, followerID, followeeID)
	if err != nil {
		return false, fmt.Errorf("delete follow edge: %w", err)
	}

	return tag.RowsAffected() == 1, nil
}

func (r *FollowRepository) ListFollowersPage(ctx context.Context, userID string, limit int32, cursor *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error) {
	baseQuery := `
SELECT follower_id, created_at
FROM follows
WHERE followee_id = $1
`
	args := []any{userID}

	if cursor != nil {
		baseQuery += ` AND (created_at < $2 OR (created_at = $2 AND follower_id < $3))`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}

	baseQuery += `
ORDER BY created_at DESC, follower_id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list followers page: %w", err)
	}
	defer rows.Close()

	userIDs := make([]string, 0, limit)
	var next *pagination.EdgeCursor

	for rows.Next() {
		var userID string
		var createdAt time.Time
		if err := rows.Scan(&userID, &createdAt); err != nil {
			return nil, nil, fmt.Errorf("scan follower row: %w", err)
		}
		userIDs = append(userIDs, userID)
		next = &pagination.EdgeCursor{CreatedAt: createdAt, ID: userID}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate follower rows: %w", err)
	}

	if int32(len(userIDs)) < limit {
		next = nil
	}

	return userIDs, next, nil
}

func (r *FollowRepository) ListFollowingPage(ctx context.Context, userID string, limit int32, cursor *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error) {
	baseQuery := `
SELECT followee_id, created_at
FROM follows
WHERE follower_id = $1
`
	args := []any{userID}

	if cursor != nil {
		baseQuery += ` AND (created_at < $2 OR (created_at = $2 AND followee_id < $3))`
		args = append(args, cursor.CreatedAt, cursor.ID)
	}

	baseQuery += `
ORDER BY created_at DESC, followee_id DESC
LIMIT $` + fmt.Sprintf("%d", len(args)+1)
	args = append(args, limit)

	rows, err := r.pool.Query(ctx, baseQuery, args...)
	if err != nil {
		return nil, nil, fmt.Errorf("list following page: %w", err)
	}
	defer rows.Close()

	userIDs := make([]string, 0, limit)
	var next *pagination.EdgeCursor

	for rows.Next() {
		var followeeID string
		var createdAt time.Time
		if err := rows.Scan(&followeeID, &createdAt); err != nil {
			return nil, nil, fmt.Errorf("scan following row: %w", err)
		}
		userIDs = append(userIDs, followeeID)
		next = &pagination.EdgeCursor{CreatedAt: createdAt, ID: followeeID}
	}

	if err := rows.Err(); err != nil {
		return nil, nil, fmt.Errorf("iterate following rows: %w", err)
	}

	if int32(len(userIDs)) < limit {
		next = nil
	}

	return userIDs, next, nil
}
