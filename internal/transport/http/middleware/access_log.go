package middleware

import (
	"net/http"
	"time"

	"go.uber.org/zap"
)

func AccessLog(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := NewStatusResponseWriter(w)

		next.ServeHTTP(rw, r)

		log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.StatusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}
