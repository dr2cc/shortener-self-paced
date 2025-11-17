package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"
)

// Get the original URL
func (h *Handler) Redirect(w http.ResponseWriter, r *http.Request) {
	uID := chi.URLParam(r, "id") //nolint:contextcheck

	shortURL, err := h.service.Expand(r.Context(), uID)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if shortURL.OriginalURL == "" {
		http.Error(w, "cant find full url", http.StatusNotFound)
		return
	}

	// if !shortURL.DeletedAt.IsZero() {
	// 	http.Error(w, "url is deleted", http.StatusGone)
	// 	return
	// }

	w.Header().Set("Content-Type", "text/html")

	http.Redirect(w, r, shortURL.OriginalURL, http.StatusTemporaryRedirect)
}
