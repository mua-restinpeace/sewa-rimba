package handler

import (
	"encoding/json"
	"errors"
	"net/http"

	"github.com/mua-restinpeace/sewa-rimba/internal/middleware"
	"github.com/mua-restinpeace/sewa-rimba/internal/repository"
	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type AuthHendler struct {
	service     *service.AuthService
	employeRepo *repository.EmployeeRepository
}

func NewAuthHandler(s *service.AuthService, employeRepo *repository.EmployeeRepository) *AuthHendler {
	return &AuthHendler{service: s, employeRepo: employeRepo}
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// POST /api/auth/login
func (h *AuthHendler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.BadRequest(w, "invalid request body")
		return
	}

	token, employee, err := h.service.Login(r.Context(), req.Email, req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		response.Unauthorized(w, "invalid email or password")
		return
	}
	if err != nil {
		response.Internal(w, "login failed")
		return
	}

	response.JSON(w, http.StatusOK, map[string]interface{}{
		"token":    token,
		"employee": employee,
	})
}

// GET /api/auth/me
func (h *AuthHendler) Me(w http.ResponseWriter, r *http.Request) {
	employeeId, ok := middleware.EmployeeIDFromContext(r.Context())
	if !ok {
		response.Unauthorized(w, "not authenticated")
		return
	}

	employee, err := h.employeRepo.GetByID(r.Context(), employeeId)
	if err != nil {
		response.NotFound(w, "employee not found")
		return
	}

	response.JSON(w, http.StatusOK, employee)
}
