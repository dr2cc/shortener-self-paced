package handler

import (
	"app/internal/lib/api/dto"
	"app/internal/lib/httpio"
	"app/internal/service"
	"encoding/json"
	"net/http"
)

// Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) принятые данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

func (h *Controller) BatchShortenAPI(w http.ResponseWriter, r *http.Request) {
	var input dto.BatchRequest

	// 1️⃣ Принимаем данные из сети, 2️⃣ десериализуем и заполняем (, &input) DTO
	if !httpio.Decode(w, r, &input) {
		return // Хелпер всё сделал за нас, просто выходим
	}

	// Проверяем заполненность URL
	if err := input.Validate(); err != nil {
		http.Error(w, "url required", http.StatusBadRequest)
		// Если хоть один URL пустой, немедленно прекращаем выполнение!
		return
	}

	// Превращаем input([]dto.RequestShortenBatch) в serviceInput ([]service.BatchInput)
	serviceInput := make([]service.BatchInput, len(input))
	for i, d := range input {
		serviceInput[i] = service.BatchInput{
			CorrelationID: d.CorrelationID,
			OriginalURL:   d.OriginalURL,
		}
	}

	// 3️⃣ Отдаем в сервис "чистые" (без json из dto) данные и получаем данные для ответа
	linkBatch, err := h.service.ShortenBatch(r.Context(), serviceInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем структуру для ответа
	output := make([]dto.ResponseShortenBatch, len(linkBatch))
	for i, shortURLBatch := range linkBatch {
		output[i] = dto.ResponseShortenBatch{
			CorrelationID: shortURLBatch.CorrelationID,
			ShortURL:      h.service.FormatShortURL(shortURLBatch.ID),
		}
	}

	// ❌TODO: переделать сериализацию на
	// httpio.Respond(

	// 4️⃣ Сериалиазуем (маршаллинг).
	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)
	if _, err = w.Write(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
