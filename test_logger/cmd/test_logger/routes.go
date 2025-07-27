package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	// Import services and handlers
)
// Handlers holds services for dependency injection
type Handlers struct {
}

// registerRoutes sets up all application routes
// This file is regenerated - do not edit manually
func registerRoutes(r *chi.Mux) {
	// Health check route (always present)
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})
	// Register handlers directly on main router (no mounting needed)
} 