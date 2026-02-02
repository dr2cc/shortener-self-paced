package handler

import (
	"app/internal/domain/link"
	"app/pkg/response"
	"encoding/json"
	"net/http"
)

// ❗Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) 1️⃣ данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

func (h *Controller) BatchShortenAPI(w http.ResponseWriter, r *http.Request) {
	// "correlation_id" и "original_url" - так в запросе, по заданию  на iter12,
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

	// 02.02.2026 Вот эту логику следует передать в service:
	// Создание batch
	// (Особенно) заполнение его (!) в цикле ниже!
	batch := make([]link.ExpandedURL, len(input))

	// ♊Как это работает вместе:
	// Handler (Эндпоинт) принимает запрос и передает данные в сервис.
	// Service (shortener) решает, что нужно создать ссылку❗
	// Link (Домен) предоставляет фабрику link.New() для сборки объекта.
	// ❌ А здесь не так!

	for i, shortURLInput := range input {
		if shortURLInput.OriginalURL == "" {
			http.Error(w, "url required", http.StatusBadRequest)
			// Если хоть один URL пустой, немедленно прекращаем выполнение!
			return
		}
		// ❌ слой handlers готовит нашу модель данных link.ExpandedURL !
		// Здесь в batch записываются все данные полученные из запроса клиента
		batch[i] = link.ExpandedURL{
			OriginalURL:   shortURLInput.OriginalURL,
			CorrelationID: shortURLInput.CorrelationID,
		}
	}

	// Передача слою сервисов.
	shortURLBatches, err := h.service.ShortenBatch(r.Context(), batch)
	// 03.02.26 В service передаем input , а уже там проводим всю работу с link.ExpandedURL
	// shortURLBatches, err := h.service.ShortenBatch(r.Context(), input)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем структуру для ответа
	res := make([]response.CreateBatchResponse, len(shortURLBatches))
	for i, shortURLBatch := range shortURLBatches {
		res[i] = response.CreateBatchResponse{
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
