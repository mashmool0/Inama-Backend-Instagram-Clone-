package service

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/user/internal/events"
)

type profileSyncStore interface {
	InsertEventIfAbsent(context.Context, pgx.Tx, string, string) (bool, error)
	CreateProfile(context.Context, pgx.Tx, string, string) error
	UpdateUsername(context.Context, pgx.Tx, string, string) error
}

type ProfileSyncProcessor struct {
	pool  *pgxpool.Pool
	store profileSyncStore
}

func NewProfileSyncProcessor(pool *pgxpool.Pool, store profileSyncStore) *ProfileSyncProcessor {
	return &ProfileSyncProcessor{pool: pool, store: store}
}

func (p *ProfileSyncProcessor) Process(ctx context.Context, body []byte) error {
	var envelope events.Envelope
	if err := json.Unmarshal(body, &envelope); err != nil {
		return fmt.Errorf("%w: decode envelope: %v", ErrRejectMessage, err)
	}
	if envelope.EventID == "" || envelope.EventType == "" {
		return fmt.Errorf("%w: missing event metadata", ErrRejectMessage)
	}
	if _, err := uuid.Parse(envelope.EventID); err != nil {
		return fmt.Errorf("%w: invalid event id", ErrRejectMessage)
	}

	var registered *events.UserRegisteredPayload
	var usernameUpdated *events.UsernameUpdatedPayload
	switch envelope.EventType {
	case events.EventTypeUserRegistered:
		var payload events.UserRegisteredPayload
		if err := decodeProfileSyncPayload(envelope.Data, &payload); err != nil {
			return err
		}
		if payload.UserID == "" || payload.Username == "" {
			return fmt.Errorf("%w: missing registration fields", ErrRejectMessage)
		}
		if _, err := uuid.Parse(payload.UserID); err != nil {
			return fmt.Errorf("%w: invalid registration user id", ErrRejectMessage)
		}
		registered = &payload
	case events.EventTypeUsernameUpdated:
		var payload events.UsernameUpdatedPayload
		if err := decodeProfileSyncPayload(envelope.Data, &payload); err != nil {
			return err
		}
		if payload.UserID == "" || payload.Username == "" {
			return fmt.Errorf("%w: missing username fields", ErrRejectMessage)
		}
		if _, err := uuid.Parse(payload.UserID); err != nil {
			return fmt.Errorf("%w: invalid username user id", ErrRejectMessage)
		}
		usernameUpdated = &payload
	default:
		return fmt.Errorf("%w: unsupported event type %s", ErrRejectMessage, envelope.EventType)
	}

	tx, err := p.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	inserted, err := p.store.InsertEventIfAbsent(ctx, tx, envelope.EventID, envelope.EventType)
	if err != nil {
		return err
	}
	if !inserted {
		return tx.Commit(ctx)
	}

	if registered != nil {
		if err := p.store.CreateProfile(ctx, tx, registered.UserID, registered.Username); err != nil {
			return err
		}
	} else {
		if err := p.store.UpdateUsername(ctx, tx, usernameUpdated.UserID, usernameUpdated.Username); err != nil {
			return err
		}
	}

	return tx.Commit(ctx)
}

func decodeProfileSyncPayload(data []byte, target any) error {
	if err := json.Unmarshal(data, target); err != nil {
		return fmt.Errorf("%w: decode payload: %v", ErrRejectMessage, err)
	}
	return nil
}

func IsRejectMessage(err error) bool {
	return errors.Is(err, ErrRejectMessage)
}
