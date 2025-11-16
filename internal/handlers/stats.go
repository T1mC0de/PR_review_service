package handlers

import (
	"net/http"

	"pr-review-service/internal/storage"
)

type StatsHandler struct {
	*BaseHandler
}

func NewStatsHandler(storage *storage.Storage) *StatsHandler {
	return &StatsHandler{
		BaseHandler: NewBaseHandler(storage),
	}
}

func (h *StatsHandler) GetStats(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "Method Not Allowed", http.StatusMethodNotAllowed)
		return
	}

	stats, err := h.storage.GetStats(r.Context())
	if err != nil {
		h.sendError(w, http.StatusInternalServerError, "INTERNAL_ERROR", "Failed to get statistics")
		return
	}

	h.sendJSON(w, http.StatusOK, stats)
}
