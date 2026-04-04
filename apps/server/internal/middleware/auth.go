package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/whitehyun/HelloTalk/apps/server/internal/response"
	"github.com/whitehyun/HelloTalk/apps/server/internal/token"
)

type contextKey string

const UserIDKey contextKey = "userID"

func Auth(secret string) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				response.Error(w, http.StatusUnauthorized, "MISSING_TOKEN", "인증 토큰이 필요합니다")
				return
			}

			parts := strings.SplitN(authHeader, " ", 2)
			if len(parts) != 2 || parts[0] != "Bearer" {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "잘못된 토큰 형식입니다")
				return
			}

			userID, err := token.ValidateAccess(parts[1], secret)
			if err != nil {
				response.Error(w, http.StatusUnauthorized, "INVALID_TOKEN", "유효하지 않은 토큰입니다")
				return
			}

			ctx := context.WithValue(r.Context(), UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// GetUserID extracts the user ID from the request context.
func GetUserID(r *http.Request) string {
	userID, _ := r.Context().Value(UserIDKey).(string)
	return userID
}
