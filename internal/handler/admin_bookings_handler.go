package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type AdminBookingHandler struct {
	service *service.BookingService
}

func NewAdminBookingHandler(service *service.BookingService) *AdminBookingHandler {
	return &AdminBookingHandler{service: service}
}

// GET /api/admin/bookings
func (h *AdminBookingHandler) GetList(w http.ResponseWriter, r *http.Request){
	status := r.URL.Query().Get("status")
	reference := r.URL.Query().Get("reference")
	phone := r.URL.Query().Get("phone")

	// default value for pagination
	pageStr := r.URL.Query().Get("page")
	page, _ := strconv.Atoi(pageStr)
	if page < 1 {
		page = 1
	} 

	limitStr := r.URL.Query().Get("limit")
	limit, _ := strconv.Atoi(limitStr)
	if limit < 1 {
		limit = 10
	}

	if limit > 100 {
		limit = 100
	}

	result, err := h.service.GetFilteredList(r.Context(), status, reference, phone, page, limit)
	if err != nil {
		response.Internal(w, "failed to fetch list bookings")
		return
	}

	response.JSON(w, http.StatusOK, result)
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

// POST /api/admin/bookings/{id}/picked_up
func (h *AdminBookingHandler) PickedUp(w http.ResponseWriter, r *http.Request) {
	id, err := h.idParam(r)
	if err != nil {
		response.BadRequest(w, "invalid booking id")
		return
	}

	booking, err := h.service.PickedUp(r.Context(), id)
	h.responseAfterTransition(w, booking, err)
}

// POST /api/admin/bookings/{id}/returned
func (h *AdminBookingHandler) Returned(w http.ResponseWriter, r *http.Request) {
	id, err := h.idParam(r)
	if err != nil {
		response.BadRequest(w, "invalid booking id")
		return
	}

	booking, err := h.service.Returned(r.Context(), id)
	h.responseAfterTransition(w, booking, err)
}

type cancelRequest struct {
	Reason string `json:"reason"`
}

// POST /api/admin/bookings/{id}/cancel
func (h *AdminBookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	id, err := h.idParam(r)
	if err != nil {
		response.BadRequest(w, "invalid booking id")
		return
	}

	var req cancelRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	reason := model.CancelReason(req.Reason)
	if reason != model.ReasonCustomerRequest && reason != model.ReasonNoShow {
		response.BadRequest(w, "reason must be 'customer_request' or 'no_show'")
		return
	}

	booking, err := h.service.Cancel(r.Context(), id, &reason)
	h.responseAfterTransition(w, booking, err)
}

func (h *AdminBookingHandler) idParam(r *http.Request) (int, error) {
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
