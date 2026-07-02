package httptransport

import (
	"net/http"
	"strconv"

	"sakeofher/internal/domain"
	"sakeofher/internal/service"
)

type PaymentHandler struct{ services *service.Services }

func NewPaymentHandler(services *service.Services) *PaymentHandler {
	return &PaymentHandler{services: services}
}

func (h *PaymentHandler) Create(w http.ResponseWriter, r *http.Request) {
	var input domain.CreatePaymentInput
	if err := DecodeJSON(r, &input); err != nil {
		WriteError(w, http.StatusBadRequest, err.Error())
		return
	}
	payment, err := h.services.Payments.CreatePayment(r.Context(), input)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusCreated, payment)
}

func (h *PaymentHandler) DevActivate(w http.ResponseWriter, r *http.Request) {
	paymentID, err := strconv.ParseInt(r.PathValue("payment_id"), 10, 64)
	if err != nil || paymentID <= 0 {
		WriteError(w, http.StatusBadRequest, "invalid payment_id")
		return
	}
	var body struct {
		ProviderPaymentID string `json:"provider_payment_id"`
	}
	_ = DecodeJSON(r, &body)

	payment, err := h.services.Payments.MarkPaidForDev(r.Context(), paymentID, body.ProviderPaymentID)
	if err != nil {
		WriteDomainError(w, err)
		return
	}
	WriteJSON(w, http.StatusOK, payment)
}
