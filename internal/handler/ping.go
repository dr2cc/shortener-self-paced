package handler

import "net/http"

// ♊ 27.01.2026
func (h *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	if err := h.service.CheckHealth(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
