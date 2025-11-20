package handlers

import (
	"app/internal/entity"
	"encoding/json"
	"io"
	"net/http"
)

// "result" - так в ответе, по заданию  на инкремент 7, называется ключ в JSON запросе
type ResponseAPI struct {
	Result string `json:"result"`
}

func apiPublicationResult(w http.ResponseWriter, h *Handler, shortURL entity.ShortURL, status int) {
	res := ResponseAPI{Result: h.service.FormatShortURL(shortURL.ID)}

	out, err := json.Marshal(res)
	// out, err := json.Marshal(h.service.FormatShortURL(shortURL.ID))
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)

	if _, err = w.Write(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Назову json post ручку ShortenAPI - обычно при помощи json создают интерфейс
func (h *Handler) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var v entity.ShortURL

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errDecode := json.NewDecoder(reader).Decode(&v); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	if v.OriginalURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	//userID := h.getUserID(r)

	shortURL, err := h.service.Shorten(r.Context(), v.OriginalURL)

	// var notUniqueErr *storage.NotUniqueURLError
	// if errors.As(err, &notUniqueErr) {
	// 	apiPublicationResult(w, h, shortURL, http.StatusConflict)
	// 	return
	// }

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiPublicationResult(w, h, shortURL, http.StatusCreated)
}

func publicationResult(w http.ResponseWriter, h *Handler, shortURL entity.ShortURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	shortenedURL := h.service.FormatShortURL(shortURL.ID)
	if _, err := w.Write([]byte(shortenedURL)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// Shorten (save.New)
func (h *Handler) ShortenText(w http.ResponseWriter, r *http.Request) {
	// Получается этого хватает, а все остальное делает
	// chi..Use(middleware.Compress ??!
	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

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
	// 	publicationResult(w, h, shortURL, http.StatusConflict)
	// 	return
	// }
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicationResult(w, h, shortURL, http.StatusCreated)
}
