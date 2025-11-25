package handlers

import (
	"app/internal/entity"
	"encoding/json"
	"net/http"
)

// ShorteningBatchResult is shortening result of batch operation.
type ShorteningBatchResult struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

func (h *Handler) BatchShortenAPI(w http.ResponseWriter, r *http.Request) {
	// "correlation_id" и "original_url" - так в запросе, по заданию  на инкремент 12,
	// называются ключи в строках JSON запроса
	type request struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}
	var input []request

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errDecode := json.NewDecoder(reader).Decode(&input); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	batch := make([]entity.ExpandedURL, len(input))

	for i, shortURLInput := range input {
		if shortURLInput.OriginalURL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			return
		}
		// Здесь в batch записываются все данные полученные из запроса клиента
		batch[i] = entity.ExpandedURL{
			OriginalURL:   shortURLInput.OriginalURL,
			CorrelationID: shortURLInput.CorrelationID,
		}
	}

	// userID := h.getUserID(r)

	// Вход в сократитель
	shortURLBatches, err := h.service.ShortenBatch(r.Context(), batch)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем структуру для ответа
	res := make([]ShorteningBatchResult, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = ShorteningBatchResult{
			CorrelationID: shortURLBatch.CorrelationID,
			ShortURL:      h.service.FormatShortURL(shortURLBatch.ID),
		}
	}

	out, err := json.Marshal(res)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(out); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
