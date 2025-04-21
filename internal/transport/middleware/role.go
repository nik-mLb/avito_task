package middleware

import (
	"net/http"
	response "github.com/nik-mLb/avito_task/internal/transport/utils"
)

// RoleMiddleware создает middleware для проверки роли
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем роль из контекста
			role, ok := r.Context().Value(roleKey).(string)
			if !ok {
				response.SendError(r.Context(), w, http.StatusInternalServerError, "Role not found in context")
				return
			}

			// Проверяем роль
			allowed := false
			for _, r := range allowedRoles {
				if role == r {
					allowed = true
					break
				}
			}

			if !allowed {
				response.SendError(r.Context(), w, http.StatusForbidden, "Insufficient permissions")
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}