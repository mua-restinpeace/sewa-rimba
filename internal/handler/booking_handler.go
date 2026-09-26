package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type BookingService interface{
	Checkout(ctx context.Context, in service.CheckoutInput) (*model.Booking, string, error)
	LookupByReferenceAndPhone(ctx context.Context, reference, phone string) (*model.Booking, error)
}

type BookingHandler struct {
	service BookingService
}

func NewBookingHandler(service BookingService) *BookingHandler {
	return &BookingHandler{service: service}
}

type checkoutItemRequest struct {
	EquipmentItemId int `json:"equipment_item_id"`
	Quantity        int `json:"quantity"`
}

type checkoutRequest struct {
	CustomerName  string                `json:"customer_name"`
	CustomerPhone string                `json:"customer_phone"`
	StartDate     string                `json:"start_date"`
	EndDate       string                `json:"end_date"`
	Items         []checkoutItemRequest `json:"items"`
}

// POST /api/bookings
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req checkoutRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.CustomerName == "" || req.CustomerPhone == "" || len(req.Items) == 0 {
		response.BadRequest(w, "customer_name, customer_phone, and at least one item are required")
		return
	}

	start, err := time.Parse("2006-01-02", req.StartDate)
	if err != nil {
		response.BadRequest(w, "start_date must be format YYYY-MM-DD")
		return
	}

	end, err := time.Parse("2006-01-02", req.EndDate)
	if err != nil {
		response.BadRequest(w, "end_date must be format YYYY-MM-DD")
		return
	}

	items := make([]service.CheckoutItem, 0, len(req.Items))
	for _, it := range req.Items {
		if it.Quantity <= 0 {
			response.BadRequest(w, "item quantity must be greater than zero")
			return
		}
		items = append(items, service.CheckoutItem{EquipmentItemID: it.EquipmentItemId, Quantity: it.Quantity})
	}

	booking, whatsappURL, err := h.service.Checkout(r.Context(), service.CheckoutInput{
		CustomerName:  req.CustomerName,
		CustomerPhone: req.CustomerPhone,
		StartDate:     start,
		EndDate:       end,
		Items:         items,
	})

	switch {
	case errors.Is(err, service.ErrInvalidDates):
		response.BadRequest(w, err.Error())
		return
	case errors.Is(err, service.ErrItemUnavailable):
		response.Conflict(w, "ITEM_UNAVAILABLE", err.Error())
		return
	case err != nil:
		response.Internal(w, "failed to create booking")
		return
	}

	response.JSON(w, http.StatusCreated, map[string]interface{}{
		"reference":    booking.Reference,
		"status":       booking.Status,
		"expires_at":   booking.ExpiresAt,
		"whatsapp_url": whatsappURL,
	})

}

func (h *BookingHandler) Lookup(w http.ResponseWriter, r *http.Request) {
	reference := r.URL.Query().Get("reference")
	phone := r.URL.Query().Get("phone")
	if reference == "" || phone == "" {
		response.BadRequest(w, "both reference and phoen are required")
		return
	}

	booking, err := h.service.LookupByReferenceAndPhone(r.Context(), reference, phone)
	if err != nil {
		response.NotFound(w, "no booking found matching that reference and phone number")
		return
	}

	response.JSON(w, http.StatusOK, booking)
}
