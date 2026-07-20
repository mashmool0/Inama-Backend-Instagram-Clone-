package service

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"github.com/mashmool0/inama/services/user/internal/events"
	"github.com/mashmool0/inama/services/user/internal/model"
)

type stubUserRepo struct {
	getByIDFn             func(context.Context, string) (model.Profile, error)
	getByUsernameFn       func(context.Context, string) (model.Profile, error)
	updateProfileFieldsFn func(context.Context, string, model.UpdateProfileInput) (model.Profile, error)
}

func (s stubUserRepo) GetByID(ctx context.Context, userID string) (model.Profile, error) {
	return s.getByIDFn(ctx, userID)
}

func (s stubUserRepo) GetByUsername(ctx context.Context, username string) (model.Profile, error) {
	return s.getByUsernameFn(ctx, username)
}

func (s stubUserRepo) UpdateProfileFields(ctx context.Context, userID string, input model.UpdateProfileInput) (model.Profile, error) {
	return s.updateProfileFieldsFn(ctx, userID, input)
}

func TestProfileManagerUpdateProfileRejectsUnauthenticated(t *testing.T) {
	t.Parallel()

	svc := NewProfileService(stubUserRepo{}, events.NopPublisher{})
	_, err := svc.UpdateProfile(context.Background(), "", model.UpdateProfileInput{})
	if !errors.Is(err, ErrUnauthenticated) {
		t.Fatalf("UpdateProfile() error = %v, want ErrUnauthenticated", err)
	}
}

func TestProfileManagerUpdateProfileRejectsUsernameChanges(t *testing.T) {
	t.Parallel()

	svc := NewProfileService(stubUserRepo{}, events.NopPublisher{})

	_, err := svc.UpdateProfile(context.Background(), "actor", model.UpdateProfileInput{
		Username: stringPtr("taken"),
	})
	var invalidArgument InvalidArgumentError
	if !errors.As(err, &invalidArgument) {
		t.Fatalf("UpdateProfile() error = %v, want InvalidArgumentError", err)
	}
}

func TestProfileManagerGetProfileMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := NewProfileService(stubUserRepo{
		getByIDFn: func(context.Context, string) (model.Profile, error) {
			return model.Profile{}, pgx.ErrNoRows
		},
		updateProfileFieldsFn: func(context.Context, string, model.UpdateProfileInput) (model.Profile, error) {
			return model.Profile{}, nil
		},
	}, events.NopPublisher{})

	_, err := svc.GetProfile(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetProfile() error = %v, want ErrNotFound", err)
	}
}

func TestProfileManagerGetProfileByUsernameMapsNotFound(t *testing.T) {
	t.Parallel()

	svc := NewProfileService(stubUserRepo{
		getByUsernameFn: func(context.Context, string) (model.Profile, error) {
			return model.Profile{}, pgx.ErrNoRows
		},
	}, events.NopPublisher{})

	_, err := svc.GetProfileByUsername(context.Background(), "missing")
	if !errors.Is(err, ErrNotFound) {
		t.Fatalf("GetProfileByUsername() error = %v, want ErrNotFound", err)
	}
}

func stringPtr(v string) *string {
	return &v
}
