package handler

import (
	"app/internal/lib/api/dto"
	"app/internal/service"
	"encoding/json"
	"net/http"
)

// ❗Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) 1️⃣ данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

func (h *Controller) BatchShortenAPI(w http.ResponseWriter, r *http.Request) {
	// «Анмаршалинг»
	var input dto.BatchRequest

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errDecode := json.NewDecoder(reader).Decode(&input); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	} // Заполнили DTO данными из сети

	// ОДНА лаконичная проверка вместо цикла
	if err := input.Validate(); err != nil {
		http.Error(w, "url required", http.StatusBadRequest)
		// Если хоть один URL пустой, немедленно прекращаем выполнение!
		return
	}

	// Превращаем []dto.RequestShortenBatch в []service.BatchInput
	serviceInput := make([]service.BatchInput, len(input))
	for i, d := range input {
		serviceInput[i] = service.BatchInput{
			OriginalURL:   d.OriginalURL,
			CorrelationID: d.CorrelationID,
		}
	}

	// Отдаем в сервис "чистые" данные
	shortURLBatches, err := h.service.ShortenBatch(r.Context(), serviceInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем структуру для ответа
	res := make([]dto.ResponseShortenBatch, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = dto.ResponseShortenBatch{
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
