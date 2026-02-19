package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) Home(w http.ResponseWriter, r *http.Request) {
	if err := app.render(w, "home", &templateData{}); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) Tracking(w http.ResponseWriter, r *http.Request) {
	trackers, err := app.Redis.GetAllURLs(r.Context())
	if err != nil {
		app.errorLog.Println("Error fetching URLs:", err)
		trackers = nil
	}

	dataMap := make(map[string]any)
	dataMap["trackers"] = trackers
	tData := &templateData{
		Data: dataMap,
	}

	if err := app.render(w, "tracking", tData); err != nil {
		app.errorLog.Println(err)
	}
}

func (app *application) TrackingItem(w http.ResponseWriter, r *http.Request) {
	tracker, err := app.Redis.GetURL(r.Context(), chi.URLParam(r, "id"))
	if err != nil {
		app.errorLog.Println("Error fetching URL:", err)
		tracker = nil
	}

	dataMap := make(map[string]any)
	dataMap["tracker"] = tracker
	tData := &templateData{
		Data: dataMap,
	}

	if err := app.render(w, "tracking-item", tData); err != nil {
		app.errorLog.Println(err)
	}
}
