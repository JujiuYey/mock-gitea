package utils

import (
	"log"
	"mockgitea/internal/config"
	"net/http"
	"strings"
	"time"
)

func LoggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		next.ServeHTTP(w, r)

		token := strings.TrimSpace(strings.TrimPrefix(r.Header.Get("Authorization"), "token"))
		if token == "" {
			token = config.DefaultFallbackToken
		}
		log.Printf("[mock-gitea] %s %s?%s token=%s remote=%s took=%s",
			r.Method,
			r.URL.Path,
			r.URL.RawQuery,
			token,
			r.RemoteAddr,
			time.Since(start).Truncate(time.Millisecond),
		)
	})
}
