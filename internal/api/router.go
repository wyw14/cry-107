package api

import (
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/wyw14/cry-107/internal/console"
)

func Router(system *System) http.Handler {
	handler := NewHandler(system)
	pages := console.NewHandler()
	router := chi.NewRouter()
	router.Use(middleware.RequestID)
	router.Use(middleware.RealIP)
	router.Use(middleware.Recoverer)
	router.Use(middleware.Timeout(5 * time.Second))
	router.Get("/healthz", handler.health)
	for _, path := range pages.Paths() {
		router.Get(path, pages.ServePage)
	}
	router.Route("/api", func(api chi.Router) {
		api.Get("/kiln/state", handler.kilnState)
		api.Post("/kiln/load", handler.requestLoad)
		api.Post("/kiln/air-proof", handler.confirmAir)
		api.Get("/burner/state", handler.burnerState)
		api.Get("/cooler/state", handler.coolerState)
		api.Post("/cooler/air", handler.coolingAir)
		api.Post("/wasteheat/air", handler.wasteHeat)
		api.Get("/incidents", handler.incidents)
	})
	return router
}
