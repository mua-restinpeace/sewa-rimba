package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/mua-restinpeace/sewa-rimba/internal/service"
	"github.com/mua-restinpeace/sewa-rimba/pkg/response"
)

type contextKey string

const EmployeeIDKey contextKey = "employee_id"

// RequireAuth validate the jwt bearer token on admin-routes and
// injects the employee id into request context so handlers can
// read it via EmployeeIDFromContext
func RequireAuth(authService *service.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			header := r.Header.Get("Authorization")
			parts := strings.SplitN(header, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Unauthorized(w, "missing or malformed Authorization header")
				return
			}

			employeeID, err := authService.ParseToken(parts[1])
			if err != nil {
				response.Unauthorized(w, "invalid or expired token")
				return
			}

			ctx := context.WithValue(r.Context(), EmployeeIDKey, employeeID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func EmployeeIDFromContext(ctx context.Context) (int, bool) {
	id, ok := ctx.Value(EmployeeIDKey).(int)
	return id, ok
}
