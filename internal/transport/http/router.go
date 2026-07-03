package httptransport

import (
	"net/http"

	"go.uber.org/zap"

	"sakeofher/internal/service"
	httpmiddleware "sakeofher/internal/transport/http/middleware"
)

func NewRouter(services *service.Services, subscriptionPathSecret string, authSuccessRedirectURL string, jwtSecret string, log *zap.Logger) http.Handler {
	mux := http.NewServeMux()

	public := NewPublicHandler(services, subscriptionPathSecret)
	bot := NewBotHandler(services)
	site := NewSiteHandler(services)
	auth := NewAuthHandler(services, authSuccessRedirectURL, jwtSecret)

	mux.HandleFunc("GET /docs", swaggerUIHandler)
	mux.HandleFunc("GET /docs/", swaggerUIHandler)
	mux.HandleFunc("GET /swagger", swaggerUIHandler)
	mux.HandleFunc("GET /docs/openapi.yaml", openAPIHandler)
	mux.HandleFunc("GET /docs/api.yaml", openAPIHandler)

	mux.HandleFunc("GET /api/v1/health", healthHandler)

	mux.HandleFunc("GET /api/v1/auth/telegram/oauth/start", auth.TelegramOAuthStart)
	mux.HandleFunc("GET /api/v1/auth/telegram/oauth/url", auth.TelegramOAuthURL)
	mux.HandleFunc("GET /api/v1/auth/telegram/oauth/callback", auth.TelegramOAuthCallback)
	mux.HandleFunc("GET /api/v1/auth/me", auth.Me)
	mux.HandleFunc("POST /api/v1/auth/logout", auth.Logout)
	mux.HandleFunc("GET /api/v1/tariffs", public.ListTariffs)
	mux.HandleFunc("GET /api/v1/subscriptions/public/{public_token}", public.GetPublicSubscription)
	mux.HandleFunc("GET /api/v1/subscriptions/path/{subscription_path}/telegram/{telegram_id}", public.GetPublicSubscriptionByTelegramID)

	mux.HandleFunc("POST /api/v1/users/telegram", bot.Start)
	mux.HandleFunc("GET /api/v1/subscriptions/by-telegram/{telegram_id}", bot.GetSubscription)

	mux.HandleFunc("GET /api/v1/config", site.GetConfig)
	mux.HandleFunc("POST /api/v1/checkout/purchase", site.CreatePurchaseCheckoutLink)
	mux.HandleFunc("POST /api/v1/checkout/renew", site.CreateRenewCheckoutLink)

	return httpmiddleware.Recovery(log, httpmiddleware.AccessLog(log, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
