package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/user/internal/model"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetByID(ctx context.Context, userID string) (model.Profile, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, username, bio, avatar_url, follower_count, following_count, created_at
FROM users
WHERE id = $1
`, userID)

	profile, err := scanProfile(row)
	if err != nil {
		return model.Profile{}, fmt.Errorf("get user by id: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) GetByUsername(ctx context.Context, username string) (model.Profile, error) {
	row := r.pool.QueryRow(ctx, `
SELECT id, username, bio, avatar_url, follower_count, following_count, created_at
FROM users
WHERE username = $1
`, username)

	profile, err := scanProfile(row)
	if err != nil {
		return model.Profile{}, fmt.Errorf("get user by username: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) ExistsByID(ctx context.Context, userID string) (bool, error) {
	var exists bool
	if err := r.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM users WHERE id = $1)`, userID).Scan(&exists); err != nil {
		return false, fmt.Errorf("exists by id: %w", err)
	}

	return exists, nil
}

func (r *UserRepository) UpdateProfileFields(ctx context.Context, userID string, input model.UpdateProfileInput) (model.Profile, error) {
	sets := make([]string, 0, 3)
	args := make([]any, 0, 4)
	argPos := 1

	if input.Bio != nil {
		sets = append(sets, fmt.Sprintf("bio = $%d", argPos))
		args = append(args, *input.Bio)
		argPos++
	}
	if input.AvatarURL != nil {
		sets = append(sets, fmt.Sprintf("avatar_url = $%d", argPos))
		args = append(args, *input.AvatarURL)
		argPos++
	}

	if len(sets) == 0 {
		return r.GetByID(ctx, userID)
	}

	sets = append(sets, "updated_at = NOW()")
	args = append(args, userID)

	query := fmt.Sprintf(`
UPDATE users
SET %s
WHERE id = $%d
RETURNING id, username, bio, avatar_url, follower_count, following_count, created_at
`, strings.Join(sets, ", "), argPos)

	profile, err := scanProfile(r.pool.QueryRow(ctx, query, args...))
	if err != nil {
		return model.Profile{}, fmt.Errorf("update profile fields: %w", err)
	}

	return profile, nil
}

func (r *UserRepository) IncrementFollowCounters(ctx context.Context, tx pgx.Tx, followerID, followeeID string) error {
	if _, err := tx.Exec(ctx, `
UPDATE users
SET following_count = following_count + 1
WHERE id = $1
`, followerID); err != nil {
		return fmt.Errorf("increment actor following_count: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE users
SET follower_count = follower_count + 1
WHERE id = $1
`, followeeID); err != nil {
		return fmt.Errorf("increment target follower_count: %w", err)
	}

	return nil
}

func (r *UserRepository) DecrementFollowCounters(ctx context.Context, tx pgx.Tx, followerID, followeeID string) error {
	if _, err := tx.Exec(ctx, `
UPDATE users
SET following_count = GREATEST(following_count - 1, 0)
WHERE id = $1
`, followerID); err != nil {
		return fmt.Errorf("decrement actor following_count: %w", err)
	}

	if _, err := tx.Exec(ctx, `
UPDATE users
SET follower_count = GREATEST(follower_count - 1, 0)
WHERE id = $1
`, followeeID); err != nil {
		return fmt.Errorf("decrement target follower_count: %w", err)
	}

	return nil
}

func scanProfile(row pgx.Row) (model.Profile, error) {
	var profile model.Profile
	if err := row.Scan(
		&profile.ID,
		&profile.Username,
		&profile.Bio,
		&profile.AvatarURL,
		&profile.FollowerCount,
		&profile.FollowingCount,
		&profile.CreatedAt,
	); err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, pgx.ErrNoRows
		}
		return model.Profile{}, err
	}

	return profile, nil
}
