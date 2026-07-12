package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/mashmool0/inama/services/user/internal/pagination"
)

type stubUserFollowStore struct {
	existsFn func(context.Context, string) (bool, error)
}

func (s stubUserFollowStore) ExistsByID(ctx context.Context, userID string) (bool, error) {
	return s.existsFn(ctx, userID)
}

func (stubUserFollowStore) IncrementFollowCounters(context.Context, pgx.Tx, string, string) error {
	return nil
}

func (stubUserFollowStore) DecrementFollowCounters(context.Context, pgx.Tx, string, string) error {
	return nil
}

type stubFollowStore struct{}

func (stubFollowStore) CreateFollow(context.Context, pgx.Tx, string, string) (bool, error) {
	return false, nil
}

func (stubFollowStore) DeleteFollow(context.Context, pgx.Tx, string, string) (bool, error) {
	return false, nil
}

func (stubFollowStore) ListFollowersPage(context.Context, string, int32, *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error) {
	return nil, nil, nil
}

func (stubFollowStore) ListFollowingPage(context.Context, string, int32, *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error) {
	return nil, nil, nil
}

func TestValidateFollowActorAndTargetRejectsSelfFollow(t *testing.T) {
	t.Parallel()

	err := validateFollowActorAndTarget("same", "same")
	var invalid InvalidArgumentError
	if !errors.As(err, &invalid) {
		t.Fatalf("validateFollowActorAndTarget() error = %v, want InvalidArgumentError", err)
	}
}

func TestFollowManagerGetFollowersRejectsInvalidCursor(t *testing.T) {
	t.Parallel()

	svc := &FollowManager{
		users: stubUserFollowStore{
			existsFn: func(context.Context, string) (bool, error) {
				return true, nil
			},
		},
		defaultLimit: 20,
		maxLimit:     100,
	}

	_, err := svc.GetFollowers(context.Background(), "user-1", 10, "not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("GetFollowers() error = %v, want ErrInvalidCursor", err)
	}
}

func TestFollowManagerGetFollowersRejectsMissingTarget(t *testing.T) {
	t.Parallel()

	svc := &FollowManager{
		users: stubUserFollowStore{
			existsFn: func(_ context.Context, userID string) (bool, error) {
				return userID != "missing", nil
			},
		},
		defaultLimit: 20,
		maxLimit:     100,
	}

	_, err := svc.GetFollowers(context.Background(), "missing", 10, "")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetFollowers() error = %v, want ErrNotFound", err)
	}
}

func TestFollowManagerFollowRejectsMissingTarget(t *testing.T) {
	t.Parallel()

	svc := &FollowManager{
		users: stubUserFollowStore{
			existsFn: func(_ context.Context, userID string) (bool, error) {
				return userID == "actor", nil
			},
		},
	}

	_, err := svc.Follow(context.Background(), "actor", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Follow() error = %v, want ErrNotFound", err)
	}
}

func TestFollowManagerUnfollowRejectsMissingTarget(t *testing.T) {
	t.Parallel()

	svc := &FollowManager{
		users: stubUserFollowStore{
			existsFn: func(_ context.Context, userID string) (bool, error) {
				return userID == "actor", nil
			},
		},
	}

	_, err := svc.Unfollow(context.Background(), "actor", "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("Unfollow() error = %v, want ErrNotFound", err)
	}
}

func TestFollowManagerNormalizeLimitCapsValues(t *testing.T) {
	t.Parallel()

	svc := &FollowManager{defaultLimit: 20, maxLimit: 100}
	if got := svc.normalizeLimit(0); got != 20 {
		t.Fatalf("normalizeLimit(0) = %d, want 20", got)
	}
	if got := svc.normalizeLimit(500); got != 100 {
		t.Fatalf("normalizeLimit(500) = %d, want 100", got)
	}
	if got := svc.normalizeLimit(50); got != 50 {
		t.Fatalf("normalizeLimit(50) = %d, want 50", got)
	}
}
