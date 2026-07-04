package httptransport

import (
	"net/http"
	"os"
	"time"

	"sakeofher/internal/gateway/remnawave"
)

type RemnawaveHandler struct{}

func NewRemnawaveHandler() *RemnawaveHandler {
	return &RemnawaveHandler{}
}

func (h *RemnawaveHandler) ListInternalSquads(w http.ResponseWriter, r *http.Request) {
	client := remnawave.NewClient(
		os.Getenv("REMNAWAVE_BASE_URL"),
		os.Getenv("REMNAWAVE_API_TOKEN"),
		15*time.Second,
	)

	squads, err := client.ListInternalSquads(r.Context())
	if err != nil {
		WriteDomainError(w, err)
		return
	}

	WriteJSON(w, http.StatusOK, squads)
}
