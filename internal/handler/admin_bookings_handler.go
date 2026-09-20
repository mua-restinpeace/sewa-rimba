package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type AdminBookingHandler struct {
	service *service.BookingService
}

func NewAdminBookingHandler(service *service.BookingService) *AdminBookingHandler {
	return &AdminBookingHandler{service: service}
}

// POST /api/admin/bookings/{id}/confirm
func (h *AdminBookingHandler) Confirm(w http.ResponseWriter, r *http.Request) {
	id, err := h.idParam(r)
	if err != nil {
		response.BadRequest(w, "invalid booking id")
		return
	}

	booking, err := h.service.Confirm(r.Context(), id)
	h.responseAfterTransition(w, booking, err)
}

func (h *AdminBookingHandler) idParam(r *http.Request) (int , error) {
	return strconv.Atoi(chi.URLParam(r, "id"))
}

func (h *AdminBookingHandler) responseAfterTransition(w http.ResponseWriter, booking interface{}, err error) {
	switch {
	case errors.Is(err, service.ErrInvalidStatus):
		response.Conflict(w, "INVALID_STATUS_TRANSITION", err.Error())
	case errors.Is(err, service.ErrItemUnavailable):
		response.Conflict(w, "ITEM_UNAVAILABLE", err.Error())
	case err != nil:
		response.NotFound(w, "booking not found")
	default:
		response.JSON(w, http.StatusOK, booking)
	}
}