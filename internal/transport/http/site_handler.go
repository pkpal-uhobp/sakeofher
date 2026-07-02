package httptransport

import (
	"net/http"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type SiteHandler struct{ services *service.Services }

func NewSiteHandler(services *service.Services) *SiteHandler {
	return &SiteHandler{services: services}
}

func (h *SiteHandler) PurchaseSubscription(w http.ResponseWriter, r *http.Request) {
	var input domain.SitePurchaseInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Subscriptions.PurchaseFromSite(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, item)
}

func (h *SiteHandler) RenewSubscription(w http.ResponseWriter, r *http.Request) {
	var input domain.SiteRenewInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Subscriptions.RenewFromSite(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
