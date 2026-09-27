package handler

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type CategoryService interface {
	Create(ctx context.Context, name string) (*model.Category, error)
	GetBySlug(ctx context.Context, slug string) (*model.Category, error)
	GetList(ctx context.Context) ([]model.Category, error)
}

type CategoryHandler struct{
	service CategoryService
}

type createRequest struct{
	Name string `json:"name"`
}

func NewCategoryHandler(service CategoryService) *CategoryHandler{
	return &CategoryHandler{service: service}
}

// POST /api/admin/categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request){
	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	if req.Name == "" {
		response.BadRequest(w, "name is required")
		return
	}

	category, err := h.service.Create(r.Context(), req.Name)
	if err != nil {
		response.Internal(w, "failed to create category")
		return
	}

	response.JSON(w, http.StatusCreated, category)
}

// GET /api/admin/categories/{slug}
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request){
	slug := chi.URLParam(r, "slug")
	category, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		response.NotFound(w, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, category)
}

// GET /api/admin/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request){
	categories, err := h.service.GetList(r.Context())
	if err != nil {
		response.Internal(w, "failed to load categories")
		return
	}

	response.JSON(w, http.StatusOK, categories)
}