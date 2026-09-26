package handler

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
)

func TestAdminBookingHandler_Confirm_Success(t *testing.T) {
	svc := &mockAdminBookingService{
		confirmFn: func(ctx context.Context, bookingId int) (*model.Booking, error) {
			if bookingId != 42 {
				t.Fatalf("expected booking id 42, got %d", bookingId)
			}
			return &model.Booking{ID: 42, Status: model.StatusConfirmed}, nil
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/confirm", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminBookingHandler_Confirm_InvalidStatusTransition(t *testing.T) {
	svc := &mockAdminBookingService{
		confirmFn: func(ctx context.Context, bookingID int) (*model.Booking, error) {
			return nil, service.ErrInvalidStatus
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/confirm", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 for an invalid status transition (e.g. confirming an already-returned booking), got %d", rec.Code)
	}
}

func TestAdminBookingHandler_Confirm_InvalidID(t *testing.T) {
	h := NewAdminBookingHandler(&mockAdminBookingService{})

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/confirm", h.Confirm)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/not-a-number/confirm", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a non-numeric id, got %d", rec.Code)
	}
}

func TestAdminBookingHandler_Cancel_Success(t *testing.T) {
	svc := &mockAdminBookingService{
		cancelFn: func(ctx context.Context, bookingID int, reason *model.CancelReason) (*model.Booking, error) {
			if *reason != model.ReasonNoShow {
				t.Fatalf("expected reason no_show, got %s", *reason)
			}
			return &model.Booking{ID: bookingID, Status: model.StatusCancelled, CancelReason: reason}, nil
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/cancel", h.Cancel)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/cancel", bytes.NewBufferString(`{"reason":"no_show"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminBookingHandler_Cancel_InvalidReason(t *testing.T) {
	h := NewAdminBookingHandler(&mockAdminBookingService{})

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/cancel", h.Cancel)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/cancel", bytes.NewBufferString(`{"reason":"changed_my_mind"}`))
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for a reason outside the allowed enum, got %d", rec.Code)
	}
}

func TestAdminBookingHandler_Pickup_Success(t *testing.T) {
	svc := &mockAdminBookingService{
		markPickedUpFn: func(ctx context.Context, bookingID int) (*model.Booking, error) {
			return &model.Booking{ID: bookingID, Status: model.StatusOngoing}, nil
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/pickup", h.PickedUp)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/pickup", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminBookingHandler_Return_Success(t *testing.T) {
	svc := &mockAdminBookingService{
		markReturnedFn: func(ctx context.Context, bookingID int) (*model.Booking, error) {
			return &model.Booking{ID: bookingID, Status: model.StatusReturned}, nil
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/return", h.Returned)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/42/return", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestAdminBookingHandler_Return_BookingNotFound(t *testing.T) {
	svc := &mockAdminBookingService{
		markReturnedFn: func(ctx context.Context, bookingID int) (*model.Booking, error) {
			return nil, errBoom
		},
	}
	h := NewAdminBookingHandler(svc)

	r := chi.NewRouter()
	r.Post("/api/admin/bookings/{id}/return", h.Returned)

	req := httptest.NewRequest(http.MethodPost, "/api/admin/bookings/999/return", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
