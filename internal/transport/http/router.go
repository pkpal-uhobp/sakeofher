package httptransport

import (
	"net/http"
	"time"

	"go.uber.org/zap"

	"sakeofher/internal/service"
)

func NewRouter(services *service.Services, log *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	public := NewPublicHandler(services)
	bot := NewBotHandler(services)
	site := NewSiteHandler(services)

	mux.HandleFunc("GET /docs", swaggerUIHandler)
	mux.HandleFunc("GET /docs/", swaggerUIHandler)
	mux.HandleFunc("GET /swagger", swaggerUIHandler)
	mux.HandleFunc("GET /docs/openapi.yaml", openAPIHandler)
	mux.HandleFunc("GET /docs/api.yaml", openAPIHandler)

	mux.HandleFunc("GET /api/v1/health", healthHandler)
	mux.HandleFunc("GET /api/v1/tariffs", public.ListTariffs)
	mux.HandleFunc("GET /api/v1/public/subscriptions/{public_token}", public.GetPublicSubscription)

	mux.HandleFunc("POST /api/v1/bot/start", bot.Start)
	mux.HandleFunc("GET /api/v1/bot/users/{telegram_id}/subscription", bot.GetSubscription)

	mux.HandleFunc("POST /api/v1/site/subscriptions/purchase", site.PurchaseSubscription)
	mux.HandleFunc("POST /api/v1/site/subscriptions/renew", site.RenewSubscription)

	return withRecovery(log, withAccessLog(log, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}

func withAccessLog(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		rw := &statusResponseWriter{ResponseWriter: w, statusCode: http.StatusOK}

		next.ServeHTTP(rw, r)

		log.Info("http request",
			zap.String("method", r.Method),
			zap.String("path", r.URL.Path),
			zap.Int("status", rw.statusCode),
			zap.Duration("duration", time.Since(start)),
			zap.String("remote_addr", r.RemoteAddr),
		)
	})
}

func withRecovery(log *zap.Logger, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if recovered := recover(); recovered != nil {
				log.Error("http panic recovered", zap.Any("panic", recovered))
				WriteJSON(w, http.StatusInternalServerError, map[string]string{"error": "internal server error"})
			}
		}()

		next.ServeHTTP(w, r)
	})
}

type statusResponseWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *statusResponseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}
