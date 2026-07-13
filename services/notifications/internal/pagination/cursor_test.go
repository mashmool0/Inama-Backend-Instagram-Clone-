package pagination

import (
	"errors"
	"testing"
	"time"
)

func TestEncodeDecodeRoundTrip(t *testing.T) {
	t.Parallel()

	original := &EdgeCursor{
		CreatedAt: time.Date(2026, 7, 12, 9, 0, 0, 123456789, time.UTC),
		ID:        "notif-123",
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
		t.Fatal("Decode() returned nil")
	}
	if decoded.ID != original.ID {
		t.Fatalf("decoded ID = %q, want %q", decoded.ID, original.ID)
	}
	if !decoded.CreatedAt.Equal(original.CreatedAt) {
		t.Fatalf("decoded CreatedAt = %v, want %v", decoded.CreatedAt, original.CreatedAt)
	}
}

func TestDecodeRejectsMalformedCursor(t *testing.T) {
	t.Parallel()

	_, err := Decode("not-base64")
	if !errors.Is(err, ErrInvalidCursor) {
		t.Fatalf("Decode() error = %v, want ErrInvalidCursor", err)
	}
}
