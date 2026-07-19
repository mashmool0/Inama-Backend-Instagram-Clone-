package pagination

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"time"
)

var ErrInvalidCursor = errors.New("invalid cursor")

type EdgeCursor struct {
	CreatedAt time.Time
	ID        string
}

type wireCursor struct {
	CreatedAt string `json:"created_at"`
	ID        string `json:"id"`
}

func Encode(cursor *EdgeCursor) (string, error) {
	if cursor == nil {
		return "", nil
	}

	payload, err := json.Marshal(wireCursor{
		CreatedAt: cursor.CreatedAt.UTC().Format(time.RFC3339Nano),
		ID:        cursor.ID,
	})
	if err != nil {
		return "", err
	}

	return base64.RawURLEncoding.EncodeToString(payload), nil
}

func Decode(encoded string) (*EdgeCursor, error) {
	if encoded == "" {
		return nil, nil
	}

	payload, err := base64.RawURLEncoding.DecodeString(encoded)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	var wire wireCursor
	if err := json.Unmarshal(payload, &wire); err != nil {
		return nil, ErrInvalidCursor
	}
	if wire.ID == "" || wire.CreatedAt == "" {
		return nil, ErrInvalidCursor
	}

	createdAt, err := time.Parse(time.RFC3339Nano, wire.CreatedAt)
	if err != nil {
		return nil, ErrInvalidCursor
	}

	return &EdgeCursor{
		CreatedAt: createdAt.UTC(),
		ID:        wire.ID,
	}, nil
}
