package handler

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/mua-restinpeace/sewa-rimba/internal/model"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type CategoryService interface {
	Create(ctx context.Context, name string) (*model.Category, error)
	GetBySlug(ctx context.Context, slug string) (*model.Category, error)
	GetList(ctx context.Context) ([]model.Category, error)
	UpdateCategory(ctx context.Context, categoryId int, name string) (*model.Category, error)
	DeleteCategory(ctx context.Context, categoryId int) error
}

type CategoryHandler struct {
	service CategoryService
}

type createRequest struct {
	Name string `json:"name"`
}

type updateRequest struct {
	Name string `json:"name"`
}

func NewCategoryHandler(service CategoryService) *CategoryHandler {
	return &CategoryHandler{service: service}
}

// POST /api/admin/categories
func (h *CategoryHandler) CreateCategory(w http.ResponseWriter, r *http.Request) {
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
func (h *CategoryHandler) Get(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	category, err := h.service.GetBySlug(r.Context(), slug)
	if err != nil {
		response.NotFound(w, "category not found")
		return
	}

	response.JSON(w, http.StatusOK, category)
}

// GET /api/admin/categories
func (h *CategoryHandler) List(w http.ResponseWriter, r *http.Request) {
	categories, err := h.service.GetList(r.Context())
	if err != nil {
		response.Internal(w, "failed to load categories")
		return
	}

	response.JSON(w, http.StatusOK, categories)
}

// PUT /api/admin/categories/{id}
func (h *CategoryHandler) Update(w http.ResponseWriter, r *http.Request) {
	catgoryIdStr := chi.URLParam(r, "id")
	categoryId, err := strconv.Atoi(catgoryIdStr)
	if err != nil {
		response.BadRequest(w, "invalid category id")
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	category, err := h.service.UpdateCategory(r.Context(), categoryId, req.Name)
	if err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			response.BadRequest(w, "category not found")
			return
		}
		response.Internal(w, "failed to update category")
		return
	}

	response.JSON(w, http.StatusOK, category)
}

// DELETE /api/admin/categories/{id}
func (h *CategoryHandler) Delete(w http.ResponseWriter, r *http.Request) {
	categoryIdStr := chi.URLParam(r, "id")
	categoryId, err := strconv.Atoi(categoryIdStr)
	if err != nil {
		response.BadRequest(w, "invalid category id")
		return
	}

	if err := h.service.DeleteCategory(r.Context(), categoryId); err != nil {
		if errors.Is(err, repository.ErrCategoryNotFound) {
			response.NotFound(w, "category not found")
			return
		}
		response.Internal(w, "failed to delete category")
		return
	}

	response.JSON(w, http.StatusNoContent, nil)
}
