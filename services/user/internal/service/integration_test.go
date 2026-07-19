package service

import (
	"context"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"github.com/mashmool0/inama/services/user/internal/events"
	"github.com/mashmool0/inama/services/user/internal/repository"
	"github.com/mashmool0/inama/services/user/internal/schema"
)

const integrationDatabaseEnv = "INTEGRATION_DATABASE_URL"

func TestFollowServiceIntegrationFollowIdempotencyAndCounters(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)
	loadSeed(t, pool)

	users := repository.NewUserRepository(pool)
	follows := repository.NewFollowRepository(pool)
	svc := NewFollowService(pool, users, follows, events.NopPublisher{}, 20, 100)

	created, err := svc.Follow(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Follow() first call error = %v", err)
	}
	if !created {
		t.Fatal("Follow() first call should create edge")
	}

	created, err = svc.Follow(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Follow() second call error = %v", err)
	}
	if created {
		t.Fatal("Follow() second call should be idempotent and not create edge")
	}

	alice, err := users.GetByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	if err != nil {
		t.Fatalf("GetByID(alice) error = %v", err)
	}
	bob, err := users.GetByID(context.Background(), "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("GetByID(bob) error = %v", err)
	}
	if alice.FollowingCount != 1 {
		t.Fatalf("alice following_count = %d, want 1", alice.FollowingCount)
	}
	if bob.FollowerCount != 1 {
		t.Fatalf("bob follower_count = %d, want 1", bob.FollowerCount)
	}
}

func TestFollowServiceIntegrationUnfollowIdempotency(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)
	loadSeed(t, pool)

	users := repository.NewUserRepository(pool)
	follows := repository.NewFollowRepository(pool)
	svc := NewFollowService(pool, users, follows, events.NopPublisher{}, 20, 100)

	if _, err := svc.Follow(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222"); err != nil {
		t.Fatalf("Follow() setup error = %v", err)
	}

	deleted, err := svc.Unfollow(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Unfollow() first call error = %v", err)
	}
	if !deleted {
		t.Fatal("Unfollow() first call should delete edge")
	}

	deleted, err = svc.Unfollow(context.Background(), "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222")
	if err != nil {
		t.Fatalf("Unfollow() second call error = %v", err)
	}
	if deleted {
		t.Fatal("Unfollow() second call should be idempotent and not delete edge")
	}

	alice, _ := users.GetByID(context.Background(), "11111111-1111-1111-1111-111111111111")
	bob, _ := users.GetByID(context.Background(), "22222222-2222-2222-2222-222222222222")
	if alice.FollowingCount != 0 || bob.FollowerCount != 0 {
		t.Fatalf("expected counters reset to zero, got alice=%d bob=%d", alice.FollowingCount, bob.FollowerCount)
	}
}

func TestFollowServiceIntegrationFollowersPagination(t *testing.T) {
	pool := integrationPool(t)
	resetDatabase(t, pool)
	loadSeed(t, pool)

	ctx := context.Background()
	_, err := pool.Exec(ctx, `
INSERT INTO follows (follower_id, followee_id, created_at)
VALUES
($1, $2, $3),
($4, $2, $5),
($6, $2, $7)
`, "11111111-1111-1111-1111-111111111111", "22222222-2222-2222-2222-222222222222", time.Date(2026, 7, 12, 12, 0, 3, 0, time.UTC),
		"33333333-3333-3333-3333-333333333333", time.Date(2026, 7, 12, 12, 0, 2, 0, time.UTC),
		"44444444-4444-4444-4444-444444444444", time.Date(2026, 7, 12, 12, 0, 1, 0, time.UTC))
	if err != nil {
		t.Fatalf("insert follows error = %v", err)
	}

	users := repository.NewUserRepository(pool)
	follows := repository.NewFollowRepository(pool)
	svc := NewFollowService(pool, users, follows, events.NopPublisher{}, 2, 2)

	page1, err := svc.GetFollowers(ctx, "22222222-2222-2222-2222-222222222222", 2, "")
	if err != nil {
		t.Fatalf("GetFollowers() page1 error = %v", err)
	}
	if len(page1.UserIDs) != 2 {
		t.Fatalf("page1 len = %d, want 2", len(page1.UserIDs))
	}
	if page1.NextCursor == "" {
		t.Fatal("page1 next_cursor should not be empty")
	}

	page2, err := svc.GetFollowers(ctx, "22222222-2222-2222-2222-222222222222", 2, page1.NextCursor)
	if err != nil {
		t.Fatalf("GetFollowers() page2 error = %v", err)
	}
	if len(page2.UserIDs) != 1 {
		t.Fatalf("page2 len = %d, want 1", len(page2.UserIDs))
	}
	if page2.NextCursor != "" {
		t.Fatalf("page2 next_cursor = %q, want empty", page2.NextCursor)
	}
}

func integrationPool(t *testing.T) *pgxpool.Pool {
	t.Helper()

	dsn := os.Getenv(integrationDatabaseEnv)
	if dsn == "" {
		t.Skipf("%s is not set", integrationDatabaseEnv)
	}

	pool, err := repository.OpenPool(context.Background(), dsn)
	if err != nil {
		t.Fatalf("OpenPool() error = %v", err)
	}
	t.Cleanup(pool.Close)

	if err := schema.Bootstrap(context.Background(), pool); err != nil {
		t.Fatalf("Bootstrap() error = %v", err)
	}

	return pool
}

func resetDatabase(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	if _, err := pool.Exec(context.Background(), `TRUNCATE TABLE follows, users`); err != nil {
		t.Fatalf("TRUNCATE error = %v", err)
	}
}

func loadSeed(t *testing.T, pool *pgxpool.Pool) {
	t.Helper()

	seedPath := filepath.Join("..", "..", "testdata", "seed.sql")
	payload, err := os.ReadFile(seedPath)
	if err != nil {
		t.Fatalf("ReadFile(%s) error = %v", seedPath, err)
	}

	if _, err := pool.Exec(context.Background(), string(payload)); err != nil {
		t.Fatalf("Exec(seed.sql) error = %v", err)
	}
}
