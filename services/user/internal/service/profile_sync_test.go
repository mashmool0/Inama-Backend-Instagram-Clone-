package service

import (
	"context"
	"errors"
	"testing"
)

func TestProfileSyncProcessorRejectsInvalidEventID(t *testing.T) {
	t.Parallel()

	processor := NewProfileSyncProcessor(nil, nil)
	err := processor.Process(context.Background(), []byte(`{
  "event_id":"not-a-uuid",
  "event_type":"user.registered",
  "data":{"user_id":"11111111-1111-1111-1111-111111111111","username":"alice"}
}`))
	if !errors.Is(err, ErrRejectMessage) {
		t.Fatalf("Process() error = %v, want ErrRejectMessage", err)
	}
}

func TestProfileSyncProcessorRejectsInvalidUserID(t *testing.T) {
	t.Parallel()

	processor := NewProfileSyncProcessor(nil, nil)
	err := processor.Process(context.Background(), []byte(`{
  "event_id":"aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa",
  "event_type":"user.registered",
  "data":{"user_id":"u1","username":"alice"}
}`))
	if !errors.Is(err, ErrRejectMessage) {
		t.Fatalf("Process() error = %v, want ErrRejectMessage", err)
	}
}
