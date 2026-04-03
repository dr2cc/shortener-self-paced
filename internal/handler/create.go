package handler

import (
	"app/internal/lib/api/dto"
	"app/internal/lib/httpio"
	err_repo "app/internal/lib/repository"
	"app/internal/service"
	"errors"
	"io"
	"net/http"
)

// Логика работы слоя http обработчиков (соответственно и каждого обработчика):
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) принятые данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

// Определяем отдаваемый статус (для ShortenAPI и ShortenText. BatchShortenAPI пока (03.04.26) не проверял )
func (h *Controller) mapErrorToStatus(err error) int {
	var notUniqueErr *err_repo.NotUniqueURLError

	switch {
	case err == nil:
		return http.StatusCreated
	case errors.As(err, &notUniqueErr):
		return http.StatusConflict
	default:
		return http.StatusInternalServerError
	}
}

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
	linkBatch, err := h.shortener.ShortenBatch(r.Context(), serviceInput)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Заполняем структуру для ответа
	output := make([]dto.ResponseShortenBatch, len(linkBatch))
	for i, shortURLBatch := range linkBatch {
		output[i] = dto.ResponseShortenBatch{
			CorrelationID: shortURLBatch.CorrelationID,
			ShortURL:      h.shortener.FormatShortURL(shortURLBatch.ID),
		}
	}

	// 4️⃣ Возвращаем клиенту response
	// Для тяжелых данных (batch ничем не ограничен) используем RespondStream
	httpio.RespondStream(w, r, http.StatusCreated, output) // успех
}

func (h *Controller) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var input dto.RequestShorten

	// 1. Вся валидация транспорта в одной строке
	// 1️⃣ Принимаем данные из сети, 2️⃣ десериализуем и заполняем (, &input) DTO
	if !httpio.Decode(w, r, &input) {
		return // Хелпер обработал все ошибки за нас, просто выходим
	}

	// 2. Валидация заполненности URL (в едином стиле JSON-ответов)
	if input.URL == "" {
		httpio.Respond(w, r, http.StatusBadRequest, map[string]string{"error": "url required"})
		return
	}

	// 3️⃣ // Пытаемся сократить URL. Просто вытаскиваем строку из DTO и отдаем сервису
	link, err := h.shortener.ShortenURL(r.Context(), input.URL)

	// Определяем статус
	status := h.mapErrorToStatus(err)

	// Если это "неизвестная" ошибка (500), отдаем http.Error и выходим
	if status == http.StatusInternalServerError {
		// Используем httpio для ошибок, чтобы сохранить формат JSON
		httpio.Respond(w, r, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}

	// Мы "отсекли" плохие ошибки.
	// Теперь основной поток кода (счастливый путь + 409) идет прямо, без вложенных if.

	// Если мы здесь, значит всё "ОК" (либо создали новый, либо нашли старый 409).
	// В обоих случаях нам нужно сформировать один и тот же ответ.
	output := dto.ResponseShorten{
		Result: h.shortener.FormatShortURL(link.ID),
	}

	// Один раз создаем output и один раз вызываем Respond. Это избавляет от ошибок при будущем изменении формата ответа.

	// 4️⃣ Возвращаем клиенту response, один раз
	httpio.Respond(w, r, status, output)
}

func (h *Controller) ShortenText(w http.ResponseWriter, r *http.Request) {
	// 1️⃣ Принимаем данные от клиента.
	// Читаем напрямую из r.Body (Middleware уже всё распаковало).
	body, err := io.ReadAll(r.Body)
	if err != nil {
		// Тестировать этот случай не будем.
		// В реальной жизни такая ошибка случается крайне редко (соединение оборвалось прямо во время передачи данных).
		// Через httptest.NewRequest получить ошибку чтения тела практически невозможно, так как bytes.Buffer или strings.Reader всегда отдают данные успешно.
		http.Error(w, "cannot read body", http.StatusBadRequest)
		return
	}
	defer r.Body.Close()

	urlStr := string(body)
	if urlStr == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// 3️⃣ Передаем данные в службу нашего приложения.
	// Сервис возвращает структуру ExpandedURL
	newLink, err := h.shortener.ShortenURL(r.Context(), urlStr) // , userID

	// Определяем статус
	status := h.mapErrorToStatus(err)

	// Если это "неизвестная" ошибка (500), отдаем http.Error и выходим
	if status == http.StatusInternalServerError {
		http.Error(w, err.Error(), status)
		return
	}

	// 4️⃣ Возвращаем клиенту response.
	// 1. Готовим данные
	content := h.shortener.FormatShortURL(newLink.ID)
	// 2. Устанавливаем заголовки
	w.Header().Set("Content-Type", "text/plain")
	// 3. Отправляем статус
	w.WriteHeader(status)
	// 4. Пишем тело (ошибку обычно игнорируют)
	_, _ = w.Write([]byte(content))
}
