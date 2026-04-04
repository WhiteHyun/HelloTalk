package middleware

import (
	"log"
	"net/http"
	"runtime/debug"

	"github.com/whitehyun/HelloTalk/apps/server/internal/response"
)

func Recovery(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				log.Printf("panic recovered: %v\n%s", err, debug.Stack())
				response.Error(w, http.StatusInternalServerError, "INTERNAL_ERROR", "서버 내부 오류가 발생했습니다")
			}
		}()
		next.ServeHTTP(w, r)
	})
}
