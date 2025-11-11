package handlers

import (
	"app/internal/entity"
	"io"
	"net/http"
)

// Shorten (save.New)
func (h *Handler) ShortenText(w http.ResponseWriter, r *http.Request) {
	// Получается этого хватает, а все остальное делает
	// chi..Use(middleware.Compress ??!
	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	//

	url, err := io.ReadAll(reader)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if string(url) == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// userID := h.getUserID(r)

	shortURL, err := h.service.Shorten(r.Context(), string(url)) // , userID

	// var notUniqueErr *storage.NotUniqueURLError
	// if errors.As(err, &notUniqueErr) {
	// 	writeShorteningResult(w, h, shortURL, http.StatusConflict)
	// 	return
	// }
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// if err = h.addEncryptedUserIDToCookie(&w, userID); err != nil {
	// 	http.Error(w, err.Error(), http.StatusInternalServerError)
	// }

	writeShorteningResult(w, h, shortURL, http.StatusCreated)
}

func writeShorteningResult(w http.ResponseWriter, h *Handler, shortURL entity.ShortURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	shortenedURL := h.service.FormatShortURL(shortURL.ID)
	if _, err := w.Write([]byte(shortenedURL)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
