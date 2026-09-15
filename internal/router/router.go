package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"github.com/mua-restinpeace/sewa-rimba/internal/handler"
	"github.com/mua-restinpeace/sewa-rimba/internal/middleware"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
)

type Handlers struct {
	Equipment *handler.EquipmentHandler
	Auth      *handler.AuthHendler
}

func New(h Handlers, authService *service.AuthService) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		// public routes
		r.Get("/equipment", h.Equipment.List)
		r.Get("/equipment/{slug}", h.Equipment.Get)

		r.Post("/auth/login", h.Auth.Login)

		// admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authService))

			r.Get("/auth/me", h.Auth.Me)
		})
	})

	return r
}
