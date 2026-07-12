package config

import "testing"

func TestNormalizePageLimits(t *testing.T) {
	t.Parallel()

	defaultLimit, maxLimit := normalizePageLimits(0, 5)
	if defaultLimit != 20 {
		t.Fatalf("defaultLimit = %d, want 20", defaultLimit)
	}
	if maxLimit != 20 {
		t.Fatalf("maxLimit = %d, want 20", maxLimit)
	}

	defaultLimit, maxLimit = normalizePageLimits(25, 100)
	if defaultLimit != 25 || maxLimit != 100 {
		t.Fatalf("normalizePageLimits(25, 100) = (%d, %d), want (25, 100)", defaultLimit, maxLimit)
	}
}
