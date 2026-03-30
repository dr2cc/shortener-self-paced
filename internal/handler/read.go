package handler

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Получаем первоначальный url, цель- перенапраление на него при получении сокращенного
func (h *Controller) Redirect(w http.ResponseWriter, r *http.Request) {
	uID := chi.URLParam(r, "id") //nolint:contextcheck

	shortURL, err := h.shortener.FindURL(r.Context(), uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if shortURL.OriginalURL == "" {
		http.Error(w, "cant find full url", http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "text/html")

	http.Redirect(w, r, shortURL.OriginalURL, http.StatusTemporaryRedirect)
}
