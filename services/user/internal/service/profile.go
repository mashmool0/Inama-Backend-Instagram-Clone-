package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mashmool0/inama/services/user/internal/model"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (model.Profile, error)
	UpdateProfile(ctx context.Context, actorID string, input model.UpdateProfileInput) (model.Profile, error)
}

type userReaderWriter interface {
	GetByID(ctx context.Context, userID string) (model.Profile, error)
	UpdateProfileFields(ctx context.Context, userID string, input model.UpdateProfileInput) (model.Profile, error)
}

type ProfileManager struct {
	users userReaderWriter
}

func NewProfileService(users userReaderWriter) *ProfileManager {
	return &ProfileManager{users: users}
}

func (s *ProfileManager) GetProfile(ctx context.Context, userID string) (model.Profile, error) {
	if userID == "" {
		return model.Profile{}, NewInvalidArgument("user_id is required")
	}

	profile, err := s.users.GetByID(ctx, userID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, ErrNotFound
		}
		return model.Profile{}, err
	}

	return profile, nil
}

func (s *ProfileManager) UpdateProfile(ctx context.Context, actorID string, input model.UpdateProfileInput) (model.Profile, error) {
	if actorID == "" {
		return model.Profile{}, ErrUnauthenticated
	}

	profile, err := s.users.UpdateProfileFields(ctx, actorID, input)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, ErrNotFound
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return model.Profile{}, ErrAlreadyExists
		}

		return model.Profile{}, err
	}

	return profile, nil
}
