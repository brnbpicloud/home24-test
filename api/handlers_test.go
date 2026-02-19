package main

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"urltracker/internal"
	"urltracker/internal/cache"
)

type mockRedisStore struct {
	storeCalled      bool
	storedTracker    *cache.URLTracker
	storeErr         error
	getURLResult     *cache.URLTracker
	getURLErr        error
	getAllURLsResult []*cache.URLTracker
	getAllURLsErr    error
}

func (m *mockRedisStore) StoreURL(ctx context.Context, tracker *cache.URLTracker) error {
	m.storeCalled = true
	m.storedTracker = tracker
	return m.storeErr
}

func (m *mockRedisStore) GetURL(ctx context.Context, id string) (*cache.URLTracker, error) {
	return m.getURLResult, m.getURLErr
}

func (m *mockRedisStore) GetAllURLs(ctx context.Context) ([]*cache.URLTracker, error) {
	return m.getAllURLsResult, m.getAllURLsErr
}

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{
			name:     "Valid HTTPS URL",
			url:      "https://example.com",
			expected: true,
		},
		{
			name:     "Valid HTTP URL",
			url:      "http://example.com",
			expected: true,
		},
		{
			name:     "Valid URL with path",
			url:      "https://example.com/path/to/page",
			expected: true,
		},
		{
			name:     "Invalid URL - no protocol",
			url:      "example.com",
			expected: false,
		},
		{
			name:     "Invalid URL - empty",
			url:      "",
			expected: false,
		},
		{
			name:     "Invalid URL - wrong protocol",
			url:      "ftp://example.com",
			expected: false,
		},
	}

	for _, tt := range tests {
		result := isValidURL(tt.url)
		if result != tt.expected {
			t.Errorf("isValidURL(%q) = %v, want %v", tt.url, result, tt.expected)
		}
	}
}

func TestSearchHandlerInvalidURL(t *testing.T) {
	app := newTestApplication()

	expectedErrorMsg := "Invalid URL. Must start with http:// or https://"

	r := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString(`{"url":"example.com"}`))
	w := httptest.NewRecorder()

	app.Search(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Search() status = %v, want %v", w.Code, http.StatusBadRequest)
	}

	var response map[string]any
	json.NewDecoder(w.Body).Decode(&response)

	if message, ok := response["message"].(string); !ok || message != expectedErrorMsg {
		t.Errorf("Search() response message = %q, want %q", response["message"], expectedErrorMsg)
	}
}

func TestSearchHandlerInvalidJSON(t *testing.T) {
	app := newTestApplication()

	r := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString(`invalid json`))
	w := httptest.NewRecorder()

	app.Search(w, r)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Search() with invalid JSON status = %v, want %v", w.Code, http.StatusBadRequest)
	}
}

func TestSearchHandlerSuccess_MockedRedis(t *testing.T) {
	app := newTestApplication()
	mockRedis := &mockRedisStore{}
	app.Redis = mockRedis

	inputURL := "https://example.com"
	r := httptest.NewRequest(http.MethodPost, "/api/search", bytes.NewBufferString(`{"url":"`+inputURL+`"}`))
	w := httptest.NewRecorder()

	app.Search(w, r)

	if w.Code != http.StatusOK {
		t.Errorf("Search() status = %v, want %v", w.Code, http.StatusOK)
	}

	if !mockRedis.storeCalled {
		t.Fatal("Search() did not call Redis StoreURL")
	}

	if mockRedis.storedTracker == nil {
		t.Fatal("Search() did not store tracker")
	}

	if mockRedis.storedTracker.URL != inputURL {
		t.Errorf("Search() stored URL = %q, want %q", mockRedis.storedTracker.URL, inputURL)
	}

	if mockRedis.storedTracker.Status != internal.StatusPending {
		t.Errorf("Search() stored Status = %q, want %q", mockRedis.storedTracker.Status, internal.StatusPending)
	}

	var response struct {
		ID    string `json:"id"`
		Error bool   `json:"error"`
	}

	if err := json.NewDecoder(w.Body).Decode(&response); err != nil {
		t.Fatalf("Search() response decode error = %v", err)
	}

	if response.Error {
		t.Error("Search() response error = true, want false")
	}

	if response.ID == "" {
		t.Error("Search() response ID is empty")
	}
}

func TestRoutes(t *testing.T) {
	app := newTestApplication()
	handler := app.routes()

	if handler == nil {
		t.Fatal("routes() returned nil handler")
	}

	tests := []struct {
		name   string
		method string
		path   string
	}{
		{
			name:   "Search endpoint CORS",
			method: http.MethodOptions,
			path:   "/api/search",
		},
		{
			name:   "Get tracking status endpoint CORS",
			method: http.MethodOptions,
			path:   "/api/tracking/123",
		},
	}

	for _, tt := range tests {
		r := httptest.NewRequest(tt.method, tt.path, nil)
		r.Header.Set("Origin", "http://example.com")
		w := httptest.NewRecorder()

		handler.ServeHTTP(w, r)

		if w.Code >= 500 {
			t.Errorf("Route %s %s returned %d, route may have issues", tt.method, tt.path, w.Code)
		}
	}
}
