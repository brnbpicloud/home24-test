package main

import (
	"errors"
	"net/http"
	"net/url"
	"strings"
	"time"
	"urltracker/internal"
	"urltracker/internal/cache"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

func (app *application) Search(w http.ResponseWriter, r *http.Request) {
	var data struct {
		URL string `json:"url"`
	}

	err := app.readJSON(r, &data)
	if err != nil {
		app.badRequest(w, err)
		return
	}

	if !isValidURL(data.URL) {
		app.errorLog.Println("Invalid URL:", data.URL)
		err := errors.New("Invalid URL. Must start with http:// or https://")
		app.badRequest(w, err)
		return
	}

	// Create URL tracker
	tracker := &cache.URLTracker{
		ID:        uuid.New().String(),
		URL:       data.URL,
		Status:    internal.StatusPending,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Store in Redis
	err = app.Redis.StoreURL(r.Context(), tracker)
	if err != nil {
		app.errorLog.Println("Error storing URL in Redis:", err)
		app.badRequest(w, err)
		return
	}

	var payload struct {
		ID    string `json:"id"`
		Error bool   `json:"error"`
	}

	payload.ID = tracker.ID
	payload.Error = false
	app.writeJSON(w, http.StatusOK, payload)
}

func (app *application) GetTrackingStatus(w http.ResponseWriter, r *http.Request) {
	id := chi.URLParam(r, "id")

	tracker, err := app.Redis.GetURL(r.Context(), id)
	if err != nil {
		app.errorLog.Println("Error retrieving URL from Redis:", err)
		app.badRequest(w, err)
		return
	}

	app.writeJSON(w, http.StatusOK, tracker)
}

func isValidURL(u string) bool {
	u = strings.TrimSpace(u)

	if !strings.HasPrefix(u, "http://") && !strings.HasPrefix(u, "https://") {
		return false
	}

	_, err := url.Parse(u)
	return err == nil
}
