package main

import (
	"context"
	"errors"
	"io"
	"log"
	"testing"
	"urltracker/internal"
	"urltracker/internal/cache"
)

type mockStore struct {
	dequeue    []*cache.URLTracker
	dequeueErr error
	updateErr  error
	updates    []*cache.URLTracker
}

func (m *mockStore) DequeueURL(ctx context.Context) (*cache.URLTracker, error) {
	if m.dequeueErr != nil {
		return nil, m.dequeueErr
	}
	if len(m.dequeue) == 0 {
		return nil, nil
	}
	item := m.dequeue[0]
	m.dequeue = m.dequeue[1:]
	return item, nil
}

func (m *mockStore) UpdateURL(ctx context.Context, t *cache.URLTracker) error {
	copy := *t
	m.updates = append(m.updates, &copy)
	return m.updateErr
}

func TestProcessNextSuccess(t *testing.T) {
	store := &mockStore{
		dequeue: []*cache.URLTracker{
			{ID: "1", URL: "https://example.com"},
		},
	}
	logger := log.New(io.Discard, "", 0)

	processed, err := processNext(context.Background(), store, func(t *cache.URLTracker) (string, error) {
		return "ok", nil
	}, logger)
	if err != nil {
		t.Fatalf("processNext() error = %v", err)
	}
	if !processed {
		t.Fatal("processNext() expected processed=true")
	}

	if len(store.updates) != 2 {
		t.Fatalf("UpdateURL calls = %d, want 2", len(store.updates))
	}

	if store.updates[0].Status != internal.StatusProcessing {
		t.Errorf("first update status = %q, want %q", store.updates[0].Status, internal.StatusProcessing)
	}

	if store.updates[1].Status != internal.StatusCompleted {
		t.Errorf("second update status = %q, want %q", store.updates[1].Status, internal.StatusCompleted)
	}
	if store.updates[1].Result != "ok" {
		t.Errorf("second update result = %q, want %q", store.updates[1].Result, "ok")
	}
}

func TestProcessNextCrawlError(t *testing.T) {
	store := &mockStore{
		dequeue: []*cache.URLTracker{
			{ID: "2", URL: "https://google.com"},
		},
	}
	logger := log.New(io.Discard, "", 0)

	processed, err := processNext(context.Background(), store, func(t *cache.URLTracker) (string, error) {
		return "", errors.New("fail to crawl")
	}, logger)
	if err != nil {
		t.Fatalf("processNext() error = %v", err)
	}
	if !processed {
		t.Fatal("processNext() expected processed=true")
	}

	if len(store.updates) != 2 {
		t.Fatalf("UpdateURL calls = %d, want 2", len(store.updates))
	}

	if store.updates[1].Status != internal.StatusFailed {
		t.Errorf("second update status = %q, want %q", store.updates[1].Status, internal.StatusFailed)
	}
	if store.updates[1].Error != "fail to crawl" {
		t.Errorf("second update error = %q, want %q", store.updates[1].Error, "fail to crawl")
	}
}

func TestProcessNextEmptyQueue(t *testing.T) {
	store := &mockStore{}
	logger := log.New(io.Discard, "", 0)

	processed, err := processNext(context.Background(), store, func(t *cache.URLTracker) (string, error) {
		return "", nil
	}, logger)
	if err != nil {
		t.Fatalf("processNext() error = %v", err)
	}
	if processed {
		t.Fatal("processNext() expected processed=false")
	}
	if len(store.updates) != 0 {
		t.Fatalf("UpdateURL calls = %d, want 0", len(store.updates))
	}
}
