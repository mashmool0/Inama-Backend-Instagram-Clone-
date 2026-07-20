package service

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"github.com/mashmool0/inama/services/user/internal/events"
	"github.com/mashmool0/inama/services/user/internal/model"
)

type ProfileService interface {
	GetProfile(ctx context.Context, userID string) (model.Profile, error)
	GetProfileByUsername(ctx context.Context, username string) (model.Profile, error)
	UpdateProfile(ctx context.Context, actorID string, input model.UpdateProfileInput) (model.Profile, error)
}

type userReaderWriter interface {
	GetByID(ctx context.Context, userID string) (model.Profile, error)
	GetByUsername(ctx context.Context, username string) (model.Profile, error)
	UpdateProfileFields(ctx context.Context, userID string, input model.UpdateProfileInput) (model.Profile, error)
}

type ProfileManager struct {
	users     userReaderWriter
	publisher events.Publisher
}

func NewProfileService(users userReaderWriter, publisher events.Publisher) *ProfileManager {
	return &ProfileManager{users: users, publisher: publisher}
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

func (s *ProfileManager) GetProfileByUsername(ctx context.Context, username string) (model.Profile, error) {
	if username == "" {
		return model.Profile{}, NewInvalidArgument("username is required")
	}

	profile, err := s.users.GetByUsername(ctx, username)
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
	if input.Username != nil {
		return model.Profile{}, NewInvalidArgument("username must be updated through auth")
	}

	before, err := s.users.GetByID(ctx, actorID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return model.Profile{}, ErrNotFound
		}
		return model.Profile{}, err
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

	if before.AvatarURL != profile.AvatarURL {
		_ = s.publisher.UserUpdated(ctx, events.UserUpdatedPayload{
			UserID:    profile.ID,
			Username:  profile.Username,
			AvatarURL: profile.AvatarURL,
		})
	}

	return profile, nil
}
