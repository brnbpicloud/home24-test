package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/cors"
)

func (app *application) routes() http.Handler {
	r := chi.NewRouter()

	r.Use(cors.Handler(cors.Options{
		AllowedOrigins: []string{"https://*", "http://*"},
		AllowedMethods: []string{"GET", "POST", "OPTIONS"},
	}))

	r.Route("/api", func(r chi.Router) {
		r.Post("/search", app.Search)
		r.Get("/tracking/{id}", app.GetTrackingStatus)
	})

	return r
}
