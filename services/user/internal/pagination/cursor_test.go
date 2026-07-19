package pagination

import (
	"errors"
	"testing"
	"time"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := &EdgeCursor{
		CreatedAt: time.Date(2026, 7, 12, 8, 30, 15, 123456789, time.UTC),
		ID:        "user-123",
	}

	encoded, err := Encode(original)
	if err != nil {
		t.Fatalf("Encode() error = %v", err)
	}

	decoded, err := Decode(encoded)
	if err != nil {
		t.Fatalf("Decode() error = %v", err)
	}

	if decoded == nil {
		t.Fatal("Decode() returned nil cursor")
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("created_at mismatch: got %v want %v", decoded.CreatedAt, original.CreatedAt)
	}
	if decoded.ID != original.ID {
		t.Fatalf("id mismatch: got %q want %q", decoded.ID, original.ID)
	}
}

func TestDecodeRejectsMalformedCursor(t *testing.T) {
	t.Parallel()

	_, err := Decode("not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("Decode() error = %v, want ErrInvalidCursor", err)
	}
}

func TestDecodeRejectsMissingFields(t *testing.T) {
	t.Parallel()

	cursor, err := Decode("eyJpZCI6IiJ9")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("Decode() error = %v, want ErrInvalidCursor", err)
	}
	if cursor != nil {
		t.Fatalf("Decode() cursor = %#v, want nil", cursor)
	}
}
