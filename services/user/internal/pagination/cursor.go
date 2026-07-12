package pagination

import "time"

type EdgeCursor struct {
	CreatedAt time.Time
	ID        string
}
