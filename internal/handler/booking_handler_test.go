package handler

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
)

func TestBookingHandler_Create_Success(t *testing.T) {
	svc := &mockBookingService{
		checkoutFn: func(ctx context.Context, in service.CheckoutInput) (*model.Booking, string, error) {
			if in.CustomerName != "Andi" {
				t.Fatalf("expected customer name 'Andi', got %q", in.CustomerName)
			}
			if len(in.Items) != 1 || in.Items[0].EquipmentItemID != 12 || in.Items[0].Quantity != 1 {
				t.Fatalf("unexpected items passed to service: %+v", in.Items)
			}
			return &model.Booking{
				Reference: "BK-0842",
				Status:    model.StatusPending,
				ExpiresAt: time.Now().Add(30 * time.Minute),
			}, "https://wa.me/62812xxxxxxx?text=...", nil
		},
	}
	h := NewBookingHandler(svc)

	body := `{
		"customer_name": "Andi",
		"customer_phone": "+62812xxxxxxx",
		"start_date": "2026-09-30",
		"end_date": "2026-10-02",
		"items": [{"equipment_item_id": 12, "quantity": 1}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d: %s", rec.Code, rec.Body.String())
	}
	var got map[string]interface{}
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if got["reference"] != "BK-0842" {
		t.Fatalf("expected reference BK-0842 in response, got %v", got["reference"])
	}
	if got["whatsapp_url"] == nil || got["whatsapp_url"] == "" {
		t.Fatalf("expected a whatsapp_url in response, got %v", got["whatsapp_url"])
	}
}

func TestBookingHandler_Create_MissingFields(t *testing.T) {
	h := NewBookingHandler(&mockBookingService{})

	// Missing customer_phone and items entirely.
	body := `{"customer_name": "Andi", "start_date": "2026-09-30", "end_date": "2026-10-02"}`
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestBookingHandler_Create_InvalidDateFormat(t *testing.T) {
	h := NewBookingHandler(&mockBookingService{})

	body := `{
		"customer_name": "Andi",
		"customer_phone": "+62812xxxxxxx",
		"start_date": "30-09-2026",
		"end_date": "2026-10-02",
		"items": [{"equipment_item_id": 12, "quantity": 1}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for malformed date, got %d", rec.Code)
	}
}

func TestBookingHandler_Create_ZeroQuantity(t *testing.T) {
	h := NewBookingHandler(&mockBookingService{})

	body := `{
		"customer_name": "Andi",
		"customer_phone": "+62812xxxxxxx",
		"start_date": "2026-09-30",
		"end_date": "2026-10-02",
		"items": [{"equipment_item_id": 12, "quantity": 0}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for zero quantity, got %d", rec.Code)
	}
}

func TestBookingHandler_Create_ItemUnavailable(t *testing.T) {
	svc := &mockBookingService{
		checkoutFn: func(ctx context.Context, in service.CheckoutInput) (*model.Booking, string, error) {
			return nil, "", service.ErrItemUnavailable
		},
	}
	h := NewBookingHandler(svc)

	body := `{
		"customer_name": "Andi",
		"customer_phone": "+62812xxxxxxx",
		"start_date": "2026-09-30",
		"end_date": "2026-10-02",
		"items": [{"equipment_item_id": 12, "quantity": 1}]
	}`
	req := httptest.NewRequest(http.MethodPost, "/api/bookings", bytes.NewBufferString(body))
	rec := httptest.NewRecorder()

	h.Create(rec, req)

	if rec.Code != http.StatusConflict {
		t.Fatalf("expected 409 when an item is unavailable, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_Lookup_Success(t *testing.T) {
	svc := &mockBookingService{
		lookupByReferenceAndPhone: func(ctx context.Context, reference, phone string) (*model.Booking, error) {
			if reference != "BK-0842" || phone != "+62812xxxxxxx" {
				t.Fatalf("unexpected lookup args: reference=%q phone=%q", reference, phone)
			}
			return &model.Booking{Reference: reference, Status: model.StatusConfirmed}, nil
		},
	}
	h := NewBookingHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/lookup?reference=BK-0842&phone=%2B62812xxxxxxx", nil)
	rec := httptest.NewRecorder()

	h.Lookup(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestBookingHandler_Lookup_MissingParams(t *testing.T) {
	h := NewBookingHandler(&mockBookingService{})

	// Only reference provided, no phone — should be rejected before
	// ever reaching the service, since both are required by design
	// (see api-endpoints.md: prevents exposing a booking via a
	// guessed/leaked reference alone).
	req := httptest.NewRequest(http.MethodGet, "/api/bookings/lookup?reference=BK-0842", nil)
	rec := httptest.NewRecorder()

	h.Lookup(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 when phone is missing, got %d", rec.Code)
	}
}

func TestBookingHandler_Lookup_NotFound(t *testing.T) {
	svc := &mockBookingService{
		lookupByReferenceAndPhone: func(ctx context.Context, reference, phone string) (*model.Booking, error) {
			return nil, errors.New("no rows")
		},
	}
	h := NewBookingHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/bookings/lookup?reference=BK-9999&phone=%2B62812xxxxxxx", nil)
	rec := httptest.NewRecorder()

	h.Lookup(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
