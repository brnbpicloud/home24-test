package cache

import (
	"context"
	"testing"
	"time"
	"urltracker/internal"

	"github.com/alicebob/miniredis/v2"
)

func newTestRedis(t *testing.T) (*RedisClient, func()) {
	t.Helper()

	mr, err := miniredis.Run()
	if err != nil {
		t.Fatalf("miniredis.Run() error = %v", err)
	}

	client := NewRedisClient(mr.Addr())
	cleanup := func() {
		_ = client.Close()
		mr.Close()
	}

	return client, cleanup
}

func TestStoreAndGetURL(t *testing.T) {
	client, cleanup := newTestRedis(t)
	defer cleanup()

	tracker := &URLTracker{
		ID:        "1",
		URL:       "https://example.com",
		Status:    internal.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	ctx := context.Background()
	if err := client.StoreURL(ctx, tracker); err != nil {
		t.Fatalf("StoreURL() error = %v", err)
	}

	got, err := client.GetURL(ctx, tracker.ID)
	if err != nil {
		t.Fatalf("GetURL() error = %v", err)
	}

	if got.ID != tracker.ID {
		t.Errorf("GetURL() ID = %q, want %q", got.ID, tracker.ID)
	}
}

func TestDequeueURL(t *testing.T) {
	client, cleanup := newTestRedis(t)
	defer cleanup()

	ctx := context.Background()
	tracker := &URLTracker{
		ID:        "queue-1",
		URL:       "https://example.com",
		Status:    internal.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	if err := client.StoreURL(ctx, tracker); err != nil {
		t.Fatalf("StoreURL() error = %v", err)
	}

	got, err := client.DequeueURL(ctx)
	if err != nil {
		t.Fatalf("DequeueURL() error = %v", err)
	}
	if got == nil {
		t.Fatal("DequeueURL() returned nil tracker")
	}
	if got.ID != tracker.ID {
		t.Errorf("DequeueURL() ID = %q, want %q", got.ID, tracker.ID)
	}
}
