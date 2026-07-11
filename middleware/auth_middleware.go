package middleware

import (
	"context"
	"net/http"
	"strings"

	apperrors "donapresentes/errors"
	"donapresentes/services"
)

func AuthMiddleware(authService *services.AuthService) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			tokenString := ""

			authHeader := r.Header.Get("Authorization")
			if authHeader != "" {
				parts := strings.Split(authHeader, " ")
				if len(parts) == 2 && parts[0] == "Bearer" {
					tokenString = parts[1]
				}
			}

			if tokenString == "" {
				cookie, err := r.Cookie("access_token")
				if err == nil {
					tokenString = cookie.Value
				}
			}

			if tokenString == "" {
				ErrorHandlerWithRequest(w, r, apperrors.NewUnauthorizedError("Token de autenticação não fornecido"), http.StatusUnauthorized)
				return
			}

			claims, err := authService.ValidateToken(tokenString)
			if err != nil {
				ErrorHandlerWithRequest(w, r, apperrors.NewUnauthorizedError("Token inválido ou expirado"), http.StatusUnauthorized)
				return
			}

			userID, ok := claims["user_id"].(float64)
			if !ok {
				ErrorHandlerWithRequest(w, r, apperrors.NewUnauthorizedError("Token inválido: user_id ausente"), http.StatusUnauthorized)
				return
			}
			username, _ := claims["username"].(string)
			role, _ := claims["role"].(string)

			var permissions []string
			if rawPerms, ok := claims["permissions"].([]interface{}); ok {
				for _, p := range rawPerms {
					if s, ok := p.(string); ok {
						permissions = append(permissions, s)
					}
				}
			}

			ctx := context.WithValue(r.Context(), "user_id", int(userID))
			ctx = context.WithValue(ctx, "username", username)
			ctx = context.WithValue(ctx, "role", role)
			ctx = context.WithValue(ctx, "permissions", permissions)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func AdminOnlyMiddleware() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			role := ""
			if roleVal := r.Context().Value("role"); roleVal != nil {
				role = roleVal.(string)
			}
			if role != "admin" {
				ErrorHandlerWithRequest(w, r, apperrors.NewForbiddenError("Acesso negado. Apenas administradores podem acessar este recurso"), http.StatusForbidden)
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

	return
}

func HasPermission(r *http.Request, permission string) bool {
	_, _, role := GetUserFromRequest(r)
	if role == "admin" {
		return true
	}
	if permsVal := r.Context().Value("permissions"); permsVal != nil {
		if perms, ok := permsVal.([]string); ok {
			for _, p := range perms {
				if p == permission {
					return true
				}
			}
		}
	}
	return false
}

func GetPermissionsFromRequest(r *http.Request) []string {
	if permsVal := r.Context().Value("permissions"); permsVal != nil {
		if perms, ok := permsVal.([]string); ok {
			result := make([]string, len(perms))
			copy(result, perms)
			return result
		}
	}
	return []string{}
}
