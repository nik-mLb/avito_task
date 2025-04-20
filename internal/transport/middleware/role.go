package middleware

import "net/http"

// RoleMiddleware создает middleware для проверки роли
func RoleMiddleware(allowedRoles ...string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем роль из контекста
			role, ok := r.Context().Value(roleKey).(string)
			if !ok {
				http.Error(w, "Role not found in context", http.StatusInternalServerError)
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
				http.Error(w, "Insufficient permissions", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		})
	}
}