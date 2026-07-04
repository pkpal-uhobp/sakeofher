package httptransport

import (
	"net/http"

	"go.uber.org/zap"

	"sakeofher/internal/service"
	httpmiddleware "sakeofher/internal/transport/http/middleware"
)

func NewRouter(
	services *service.Services,
	subscriptionPathSecret string,
	jwtSecret string,
	log *zap.Logger,
) http.Handler {
	mux := http.NewServeMux()

	public := NewPublicHandler(services, subscriptionPathSecret)
	users := NewUserHandler(services)
	subscriptions := NewSubscriptionAdminHandler(services)
	tariffs := NewTariffAdminHandler(services)

	bot := NewBotHandler(services)
	site := NewSiteHandler(services)
	auth := NewAuthHandler(services, jwtSecret)

	requireAdmin := func(handler http.HandlerFunc) http.Handler {
		return httpmiddleware.RequireAdmin(jwtSecret, handler)
	}

	mux.HandleFunc("GET /docs", swaggerUIHandler)
	mux.HandleFunc("GET /docs/", swaggerUIHandler)
	mux.HandleFunc("GET /swagger", swaggerUIHandler)
	mux.HandleFunc("GET /docs/openapi.yaml", openAPIHandler)
	mux.HandleFunc("GET /docs/api.yaml", openAPIHandler)

	mux.HandleFunc("GET /api/v1/health", healthHandler)

	mux.HandleFunc("POST /api/v1/auth/login", auth.Login)
	mux.HandleFunc("GET /api/v1/auth/me", auth.Me)
	mux.HandleFunc("POST /api/v1/auth/logout", auth.Logout)

	mux.HandleFunc("GET /api/v1/tariffs", public.ListTariffs)
	mux.HandleFunc("GET /api/v1/config", site.GetConfig)
	mux.HandleFunc("POST /api/v1/checkout/purchase", site.CreatePurchaseCheckoutLink)
	mux.HandleFunc("POST /api/v1/checkout/renew", site.CreateRenewCheckoutLink)

	mux.HandleFunc("GET /api/v1/subscriptions/public/{public_token}", public.GetPublicSubscription)
	mux.HandleFunc("GET /api/v1/subscriptions/path/{subscription_path}/telegram/{telegram_id}", public.GetPublicSubscriptionByTelegramID)
	mux.HandleFunc("GET /api/v1/subscriptions/by-telegram/{telegram_id}", bot.GetSubscription)

	mux.HandleFunc("POST /api/v1/users/telegram", bot.Start)

	mux.Handle("GET /api/v1/users", requireAdmin(users.List))
	mux.Handle("GET /api/v1/users/{id}", requireAdmin(users.GetByID))
	mux.Handle("PATCH /api/v1/users/{id}", requireAdmin(users.Update))
	mux.Handle("POST /api/v1/users/{id}/block", requireAdmin(users.Block))
	mux.Handle("POST /api/v1/users/{id}/unblock", requireAdmin(users.Unblock))
	mux.Handle("POST /api/v1/users/{id}/delete", requireAdmin(users.Delete))

	mux.Handle("GET /api/v1/subscriptions", requireAdmin(subscriptions.List))
	mux.Handle("GET /api/v1/subscriptions/{id}", requireAdmin(subscriptions.GetByID))
	mux.Handle("POST /api/v1/subscriptions", requireAdmin(subscriptions.CreateManual))
	mux.Handle("PATCH /api/v1/subscriptions/{id}", requireAdmin(subscriptions.Update))
	mux.Handle("POST /api/v1/subscriptions/{id}/extend", requireAdmin(subscriptions.Extend))
	mux.Handle("PATCH /api/v1/subscriptions/{id}/traffic-limit", requireAdmin(subscriptions.UpdateTrafficLimit))
	mux.Handle("POST /api/v1/subscriptions/{id}/disable", requireAdmin(subscriptions.Disable))
	mux.Handle("POST /api/v1/subscriptions/{id}/enable", requireAdmin(subscriptions.Enable))
	mux.Handle("POST /api/v1/subscriptions/{id}/cancel", requireAdmin(subscriptions.Cancel))

	mux.Handle("GET /api/v1/tariffs/all", requireAdmin(tariffs.ListAll))
	mux.Handle("GET /api/v1/tariffs/{id}", requireAdmin(tariffs.GetByID))
	mux.Handle("POST /api/v1/tariffs", requireAdmin(tariffs.Create))
	mux.Handle("PATCH /api/v1/tariffs/{id}", requireAdmin(tariffs.Update))
	mux.Handle("POST /api/v1/tariffs/{id}/enable", requireAdmin(tariffs.Enable))
	mux.Handle("POST /api/v1/tariffs/{id}/disable", requireAdmin(tariffs.Disable))

	return httpmiddleware.Recovery(log, httpmiddleware.AccessLog(log, mux))
}

func healthHandler(w http.ResponseWriter, r *http.Request) {
	WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
}
