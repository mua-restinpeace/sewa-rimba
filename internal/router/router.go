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
	Equipment    *handler.EquipmentHandler
	Booking      *handler.BookingHandler
	Auth         *handler.AuthHendler
	AdminBooking *handler.AdminBookingHandler
}

func New(h Handlers, authService *service.AuthService) http.Handler {
	r := chi.NewRouter()

	r.Use(chimiddleware.Logger)
	r.Use(chimiddleware.Recoverer)

	r.Route("/api", func(r chi.Router) {
		// public routes
		r.Get("/equipment", h.Equipment.List)
		r.Get("/equipment/{slug}", h.Equipment.Get)
		r.Post("/bookings", h.Booking.Create)

		r.Post("/auth/login", h.Auth.Login)

		// admin routes
		r.Group(func(r chi.Router) {
			r.Use(middleware.RequireAuth(authService))

			r.Get("/auth/me", h.Auth.Me)

			r.Route("/admin/bookings", func(r chi.Router) {
				r.Post("/{id}/confirm", h.AdminBooking.Confirm)
				r.Post("/{id}/pickup", h.AdminBooking.PickedUp)
				r.Post("/{id}/return", h.AdminBooking.Returned)
				r.Post("/{id}/cancel", h.AdminBooking.Cancel)
			})
		})
	})

	return r
}
