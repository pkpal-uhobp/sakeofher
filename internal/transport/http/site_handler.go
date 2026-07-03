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

func (h *SiteHandler) GetConfig(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.services.Site.GetConfig(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, cfg)
}

func (h *SiteHandler) CreatePurchaseCheckoutLink(w http.ResponseWriter, r *http.Request) {
	var input domain.SitePurchaseLinkInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Site.CreatePurchaseLink(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}

func (h *SiteHandler) CreateRenewCheckoutLink(w http.ResponseWriter, r *http.Request) {
	var input domain.SiteRenewLinkInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}

	item, err := h.services.Site.CreateRenewLink(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, item)
}
