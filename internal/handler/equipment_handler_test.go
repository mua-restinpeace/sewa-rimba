package handler

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
)

func TestEquipmentHandler_List_Success(t *testing.T) {
	svc := &mockEquipmentService{
		listAvailableFn: func(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
			return []model.EquipmentAvailability{
				{
					EquipmentItem:     model.EquipmentItem{ID: 1, Name: "Tent", TotalQuantity: 5},
					AvailableQuantity: 3,
				},
			}, nil
		},
	}
	h := NewEquipmentHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/equipment?start_date=2026-09-01&end_date=2026-09-03", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
	var got []model.EquipmentAvailability
	if err := json.Unmarshal(rec.Body.Bytes(), &got); err != nil {
		t.Fatalf("failed to decode response: %v", err)
	}
	if len(got) != 1 || got[0].AvailableQuantity != 3 {
		t.Fatalf("unexpected response body: %+v", got)
	}
}

func TestEquipmentHandler_List_MissingStartDate(t *testing.T) {
	h := NewEquipmentHandler(&mockEquipmentService{})

	req := httptest.NewRequest(http.MethodGet, "/api/equipment?end_date=2026-09-03", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rec.Code)
	}
}

func TestEquipmentHandler_List_InvalidCategoryID(t *testing.T) {
	h := NewEquipmentHandler(&mockEquipmentService{})

	req := httptest.NewRequest(http.MethodGet, "/api/equipment?start_date=2026-09-01&end_date=2026-09-03&category_id=abc", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400 for non-integer category_id, got %d", rec.Code)
	}
}

func TestEquipmentHandler_List_ServiceError(t *testing.T) {
	svc := &mockEquipmentService{
		listAvailableFn: func(ctx context.Context, start, end time.Time, categoryID *int) ([]model.EquipmentAvailability, error) {
			return nil, errBoom
		},
	}
	h := NewEquipmentHandler(svc)

	req := httptest.NewRequest(http.MethodGet, "/api/equipment?start_date=2026-09-01&end_date=2026-09-03", nil)
	rec := httptest.NewRecorder()

	h.List(rec, req)

	if rec.Code != http.StatusInternalServerError {
		t.Fatalf("expected 500 when the service errors, got %d", rec.Code)
	}
}

func TestEquipmentHandler_Get_Success(t *testing.T) {
	svc := &mockEquipmentService{
		getBySlugFn: func(ctx context.Context, slug string) (*model.EquipmentItem, error) {
			if slug != "4-person-tent" {
				t.Fatalf("expected slug '4-person-tent', got %q", slug)
			}
			return &model.EquipmentItem{ID: 1, Name: "4-Person Tent", Slug: slug}, nil
		},
	}
	h := NewEquipmentHandler(svc)

	// Registered through a real chi router so chi.URLParam resolves,
	// same as it would in production.
	r := chi.NewRouter()
	r.Get("/api/equipment/{slug}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/equipment/4-person-tent", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d: %s", rec.Code, rec.Body.String())
	}
}

func TestEquipmentHandler_Get_NotFound(t *testing.T) {
	svc := &mockEquipmentService{
		getBySlugFn: func(ctx context.Context, slug string) (*model.EquipmentItem, error) {
			return nil, errBoom
		},
	}
	h := NewEquipmentHandler(svc)

	r := chi.NewRouter()
	r.Get("/api/equipment/{slug}", h.Get)

	req := httptest.NewRequest(http.MethodGet, "/api/equipment/does-not-exist", nil)
	rec := httptest.NewRecorder()
	r.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rec.Code)
	}
}
