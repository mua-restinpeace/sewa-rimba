package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type EquipmentHandler struct {
	service *service.EquipmentService
}

func NewEquipmentHandler(s *service.EquipmentService) *EquipmentHandler {
	return &EquipmentHandler{service: s}
}

// GET /api/equipment?start_date=&end_date=&category_id=
func (h *EquipmentHandler) List(w http.ResponseWriter, r *http.Request) {
	start, err := time.Parse("2006-01-02", r.URL.Query().Get("start_date"))
	if err != nil {
		response.BadRequest(w, "start_date is required, format YYYY-MM-DD")
		return
	}

	end, err := time.Parse("2006-01-02", r.URL.Query().Get("end_date"))
	if err != nil {
		response.BadRequest(w, "end_date is required, format YYYY-MM-DD")
		return
	}

	var categoryID *int
	if raw := r.URL.Query().Get("category_id"); raw != "" {
		id, err := strconv.Atoi(raw)
		if err != nil {
			response.BadRequest(w, "category_id must be an integer")
			return
		}

		categoryID = &id
	}

	items, err := h.service.ListAvailable(r.Context(), start, end, categoryID)
	if err != nil {
		response.Internal(w, "failed to load equipments")
		return
	}

	response.JSON(w, http.StatusOK, items)
}

// GET /api/equipment/$slug
func (h *EquipmentHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	item, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		response.NotFound(w, "equipment not found")
		return
	}

	response.JSON(w, http.StatusOK, item)
}
