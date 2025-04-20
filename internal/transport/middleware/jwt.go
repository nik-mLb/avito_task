package middleware

import (
	"context"
	"net/http"

	"github.com/nik-mLb/avito_task/internal/transport/jwt"
)

// Ключи для хранения в контексте
type contextKey string

const (
	userIDKey contextKey = "userID"
	roleKey   contextKey = "role"
)

// AuthMiddleware создает middleware для проверки аутентификации
func AuthMiddleware(tokenator *jwt.Tokenator) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Получаем токен из куки
			cookie, err := r.Cookie("token")
			if err != nil {
				http.Error(w, "Token cookie is required", http.StatusUnauthorized)
				return
			}

			// Парсим токен
			claims, err := tokenator.ParseJWT(cookie.Value)
			if err != nil {
				http.Error(w, "Invalid token", http.StatusUnauthorized)
				return
			}

			// Добавляем данные в контекст
			ctx := context.WithValue(r.Context(), userIDKey, claims.UserID)
			ctx = context.WithValue(ctx, roleKey, claims.Role)

			// Передаем запрос дальше
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}