package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Get("/", app.Home)
	r.Get("/tracking", app.Tracking)
	r.Get("/tracking/{id}", app.TrackingItem)
	return r
}
