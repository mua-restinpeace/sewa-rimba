package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mua-restinpeace/sewa-rimba/internal/handler"
)

type Handlers struct {
	Equipment *handler.EquipmentHandler
}

func New(h Handlers) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		// public routes
		r.Get("/equipment", h.Equipment.List)
	})

	return r
}
