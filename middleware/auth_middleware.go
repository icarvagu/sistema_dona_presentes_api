package middleware

import (
	"context"
	"net/http"
	"strconv"
	"strings"

	apperrors "donapresentes/errors"
	"donapresentes/services"
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {

			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				ErrorHandler(w, apperrors.NewValidationError("Token de autenticação não fornecido"), http.StatusUnauthorized)
				return
			}

			parts := strings.Split(authHeader, " ")
			if len(parts) != 2 || parts[0] != "Bearer" {
				ErrorHandler(w, apperrors.NewValidationError("Formato de token inválido"), http.StatusUnauthorized)
				return
			}

			tokenString := parts[1]

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				ErrorHandler(w, apperrors.NewValidationError("Token inválido ou expirado"), http.StatusUnauthorized)
				return
			}

			userID, _ := claims["user_id"].(float64)
			username, _ := claims["username"].(string)
			role, _ := claims["role"].(string)

			ctx := context.WithValue(r.Context(), "user_id", int(userID))
			ctx = context.WithValue(ctx, "username", username)
			ctx = context.WithValue(ctx, "role", role)

			r.Header.Set("X-User-ID", strconv.Itoa(int(userID)))
			r.Header.Set("X-Username", username)
			r.Header.Set("X-User-Role", role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnlyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := r.Header.Get("X-User-Role")
			if role != "admin" {
				ErrorHandler(w, apperrors.NewValidationError("Acesso negado. Apenas administradores podem acessar este recurso"), http.StatusForbidden)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
}

func GetUserFromRequest(r *http.Request) (userID int, username string, role string) {

	if userIDVal := r.Context().Value("user_id"); userIDVal != nil {
		userID = userIDVal.(int)
	}
	if usernameVal := r.Context().Value("username"); usernameVal != nil {
		username = usernameVal.(string)
	}
	if roleVal := r.Context().Value("role"); roleVal != nil {
		role = roleVal.(string)
	}

	if userID == 0 {
		userIDStr := r.Header.Get("X-User-ID")
		if userIDStr != "" {
			if id, err := strconv.Atoi(userIDStr); err == nil {
				userID = id
			}
		}
	}
	if username == "" {
		username = r.Header.Get("X-Username")
	}
	if role == "" {
		role = r.Header.Get("X-User-Role")
	}
	return
}
