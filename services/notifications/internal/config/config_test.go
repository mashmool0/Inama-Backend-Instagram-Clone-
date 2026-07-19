package config

import (
	"reflect"
	"testing"
)

func TestParseRoutingKeysDropsWhitespaceAndEmptyValues(t *testing.T) {
	t.Parallel()

	got := parseRoutingKeys(" post.liked, , comment.created ,user.followed ")
	want := []string{"post.liked", "comment.created", "user.followed"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("parseRoutingKeys() = %#v, want %#v", got, want)
	}
}

func TestNormalizePageLimitsAppliesDefaultsAndBounds(t *testing.T) {
	t.Parallel()

	defaultLimit, maxLimit := normalizePageLimits(0, 5)
	if defaultLimit != 20 || maxLimit != 20 {
		t.Fatalf("normalizePageLimits(0, 5) = (%d, %d), want (20, 20)", defaultLimit, maxLimit)
	}

	defaultLimit, maxLimit = normalizePageLimits(30, 10)
	if defaultLimit != 30 || maxLimit != 30 {
		t.Fatalf("normalizePageLimits(30, 10) = (%d, %d), want (30, 30)", defaultLimit, maxLimit)
	}
}
