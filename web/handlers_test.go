package main

import (
	"context"
	"io"
	"log"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
	"urltracker/internal/cache"

	"github.com/go-chi/chi/v5"
)

type mockRedisStore struct {
	allURLs    []*cache.URLTracker
	getAllErr  error
	tracker    *cache.URLTracker
	getErr     error
	getAllHits int
	getHits    int
}

func (m *mockRedisStore) GetAllURLs(_ context.Context) ([]*cache.URLTracker, error) {
	m.getAllHits++
	return m.allURLs, m.getAllErr
}

func (m *mockRedisStore) GetURL(_ context.Context, _ string) (*cache.URLTracker, error) {
	m.getHits++
	return m.tracker, m.getErr
}

func newTestApplication(redis RedisStore) *application {
	return &application{
		ApiAddr:  "http://localhost:4001",
		infoLog:  log.New(io.Discard, "", 0),
		errorLog: log.New(io.Discard, "", 0),
		Redis:    redis,
	}
}

func TestHomeHandler(t *testing.T) {
	app := newTestApplication(nil)

	r := httptest.NewRequest(http.MethodGet, "/", nil)
	w := httptest.NewRecorder()

	app.Home(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Home() status = %d, want %d", w.Code, http.StatusOK)
	}

	body := w.Body.String()
	if body == "" {
		t.Fatal("Home() response body is empty")
	}
	if !strings.Contains(body, "URL Analyzer") {
		t.Errorf("Home() response missing title, got %q", body)
	}
}

func TestTrackingHandler(t *testing.T) {
	tracker := &cache.URLTracker{
		ID:        "1",
		URL:       "https://example.com",
		Status:    "pending",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRedis := &mockRedisStore{allURLs: []*cache.URLTracker{tracker}}
	app := newTestApplication(mockRedis)

	r := httptest.NewRequest(http.MethodGet, "/tracking", nil)
	w := httptest.NewRecorder()

	app.Tracking(w, r)

	if w.Code != http.StatusOK {
		t.Fatalf("Tracking() status = %d, want %d", w.Code, http.StatusOK)
	}

	if mockRedis.getAllHits != 1 {
		t.Fatalf("Tracking() GetAllURLs calls = %d, want 1", mockRedis.getAllHits)
	}

	body := w.Body.String()
	if !strings.Contains(body, "URL Tracking Dashboard") {
		t.Errorf("Tracking() response missing title")
	}
	if !strings.Contains(body, tracker.URL) {
		t.Errorf("Tracking() response missing tracker URL")
	}
}

func TestTrackingItemHandler(t *testing.T) {
	tracker := &cache.URLTracker{
		ID:        "abc",
		URL:       "https://example.com",
		Status:    "completed",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	mockRedis := &mockRedisStore{tracker: tracker}
	app := newTestApplication(mockRedis)

	router := chi.NewRouter()
	router.Get("/tracking/{id}", app.TrackingItem)

	r := httptest.NewRequest(http.MethodGet, "/tracking/abc", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, r)

	body := w.Body.String()
	if !strings.Contains(body, "URL Tracking Details") {
		t.Errorf("TrackingItem() response missing title")
	}
	if !strings.Contains(body, tracker.URL) {
		t.Errorf("TrackingItem() response missing tracker URL")
	}
}
