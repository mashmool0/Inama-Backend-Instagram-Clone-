package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/user/internal/model"
	"github.com/mashmool0/inama/services/user/internal/pagination"
)

type FollowService interface {
	Follow(ctx context.Context, actorID, targetUserID string) (bool, error)
	Unfollow(ctx context.Context, actorID, targetUserID string) (bool, error)
	GetFollowers(ctx context.Context, userID string, limit int32, cursor string) (model.UserIDPage, error)
	GetFollowing(ctx context.Context, userID string, limit int32, cursor string) (model.UserIDPage, error)
}

type userFollowStore interface {
	ExistsByID(ctx context.Context, userID string) (bool, error)
	IncrementFollowCounters(ctx context.Context, tx pgx.Tx, followerID, followeeID string) error
	DecrementFollowCounters(ctx context.Context, tx pgx.Tx, followerID, followeeID string) error
}

type followStore interface {
	CreateFollow(ctx context.Context, tx pgx.Tx, followerID, followeeID string) (bool, error)
	DeleteFollow(ctx context.Context, tx pgx.Tx, followerID, followeeID string) (bool, error)
	ListFollowersPage(ctx context.Context, userID string, limit int32, cursor *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error)
	ListFollowingPage(ctx context.Context, userID string, limit int32, cursor *pagination.EdgeCursor) ([]string, *pagination.EdgeCursor, error)
}

type FollowManager struct {
	pool         *pgxpool.Pool
	users        userFollowStore
	follows      followStore
	defaultLimit int32
	maxLimit     int32
}

func NewFollowService(pool *pgxpool.Pool, users userFollowStore, follows followStore, defaultLimit, maxLimit int32) *FollowManager {
	return &FollowManager{
		pool:         pool,
		users:        users,
		follows:      follows,
		defaultLimit: defaultLimit,
		maxLimit:     maxLimit,
	}
}

func (s *FollowManager) Follow(ctx context.Context, actorID, targetUserID string) (bool, error) {
	if err := validateFollowActorAndTarget(actorID, targetUserID); err != nil {
		return false, err
	}

	if err := s.ensureUsersExist(ctx, actorID, targetUserID); err != nil {
		return false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	created, err := s.follows.CreateFollow(ctx, tx, actorID, targetUserID)
	if err != nil {
		return false, err
	}
	if created {
		if err := s.users.IncrementFollowCounters(ctx, tx, actorID, targetUserID); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return created, nil
}

func (s *FollowManager) Unfollow(ctx context.Context, actorID, targetUserID string) (bool, error) {
	if err := validateFollowActorAndTarget(actorID, targetUserID); err != nil {
		return false, err
	}

	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return false, err
	}
	defer tx.Rollback(ctx)

	deleted, err := s.follows.DeleteFollow(ctx, tx, actorID, targetUserID)
	if err != nil {
		return false, err
	}
	if deleted {
		if err := s.users.DecrementFollowCounters(ctx, tx, actorID, targetUserID); err != nil {
			return false, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return false, err
	}

	return deleted, nil
}

func (s *FollowManager) GetFollowers(ctx context.Context, userID string, limit int32, cursor string) (model.UserIDPage, error) {
	if userID == "" {
		return model.UserIDPage{}, NewInvalidArgument("user_id is required")
	}
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return model.UserIDPage{}, err
	}

	decoded, err := pagination.Decode(cursor)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return model.UserIDPage{}, ErrInvalidCursor
		}
		return model.UserIDPage{}, err
	}

	userIDs, next, err := s.follows.ListFollowersPage(ctx, userID, s.normalizeLimit(limit), decoded)
	if err != nil {
		return model.UserIDPage{}, err
	}

	nextCursor, err := pagination.Encode(next)
	if err != nil {
		return model.UserIDPage{}, err
	}

	return model.UserIDPage{UserIDs: userIDs, NextCursor: nextCursor}, nil
}

func (s *FollowManager) GetFollowing(ctx context.Context, userID string, limit int32, cursor string) (model.UserIDPage, error) {
	if userID == "" {
		return model.UserIDPage{}, NewInvalidArgument("user_id is required")
	}
	if err := s.ensureUserExists(ctx, userID); err != nil {
		return model.UserIDPage{}, err
	}

	decoded, err := pagination.Decode(cursor)
	if err != nil {
		if errors.Is(err, pagination.ErrInvalidCursor) {
			return model.UserIDPage{}, ErrInvalidCursor
		}
		return model.UserIDPage{}, err
	}

	userIDs, next, err := s.follows.ListFollowingPage(ctx, userID, s.normalizeLimit(limit), decoded)
	if err != nil {
		return model.UserIDPage{}, err
	}

	nextCursor, err := pagination.Encode(next)
	if err != nil {
		return model.UserIDPage{}, err
	}

	return model.UserIDPage{UserIDs: userIDs, NextCursor: nextCursor}, nil
}

func (s *FollowManager) normalizeLimit(limit int32) int32 {
	if limit <= 0 {
		return s.defaultLimit
	}
	if limit > s.maxLimit {
		return s.maxLimit
	}
	return limit
}

func (s *FollowManager) ensureUsersExist(ctx context.Context, actorID, targetUserID string) error {
	if err := s.ensureUserExists(ctx, actorID); err != nil {
		return err
	}
	return s.ensureUserExists(ctx, targetUserID)
}

func (s *FollowManager) ensureUserExists(ctx context.Context, userID string) error {
	exists, err := s.users.ExistsByID(ctx, userID)
	if err != nil {
		return err
	}
	if !exists {
		return ErrNotFound
	}
	return nil
}

func validateFollowActorAndTarget(actorID, targetUserID string) error {
	if actorID == "" {
		return ErrUnauthenticated
	}
	if targetUserID == "" {
		return NewInvalidArgument("target_user_id is required")
	}
	if actorID == targetUserID {
		return NewInvalidArgument("cannot follow yourself")
	}
	return nil
}
