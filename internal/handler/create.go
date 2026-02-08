package handler

import (
	"app/internal/domain/link"
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

	// 4️⃣ Возвращаем клиенту response
	// Для тяжелых данных (batch ничем не ограничен) используем стрим
	httpio.RespondStream(w, r, http.StatusCreated, output) // успех
}

func (h *Controller) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var input dto.RequestShorten
	var notUniqueErr *err_repo.NotUniqueURLError

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
	link, err := h.service.ShortenURL(r.Context(), input.URL)

	// 2. Сначала проверяем фатальные ошибки (БД упала, сеть пропала и т.д.)
	// ЕСЛИ ошибка есть И это НЕ статус 409
	if err != nil && !errors.As(err, &notUniqueErr) {
		// Используем ваш httpio для ошибок, чтобы сохранить формат JSON (если нужно)
		httpio.Respond(w, r, http.StatusInternalServerError, map[string]string{"error": err.Error()})
		return
	}
	// Мы сначала "отсекли" плохие ошибки. Теперь основной поток кода (счастливый путь + 409) идет прямо, без вложенных if.

	// 3. Если мы здесь, значит всё "ОК" (либо создали новый, либо нашли старый 409).
	// В обоих случаях нам нужно сформировать один и тот же ответ.
	output := dto.ResponseShorten{
		Result: h.service.FormatShortURL(link.ID),
	}

	// Один раз создаем output и один раз вызываем Respond. Это избавляет от ошибок при будущем изменении формата ответа.

	// 4. Выбираем статус-код
	status := http.StatusCreated // 201 по умолчанию
	if errors.As(err, &notUniqueErr) {
		status = http.StatusConflict // 409 если не уникален
	}

	// 4️⃣ Возвращаем клиенту response, один раз
	httpio.Respond(w, r, status, output)

}

func publicationResult(w http.ResponseWriter, h *Controller, expandedURL link.ExpandedURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	content := h.service.FormatShortURL(expandedURL.ID)
	if _, err := w.Write([]byte(content)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

func (h *Controller) ShortenText(w http.ResponseWriter, r *http.Request) {
	// 1️⃣ Принимаем данные от клиента
	reader, err := httpio.GetDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	url, err := io.ReadAll(reader)
	// не получилось прочитать reader
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// пустой url
	if string(url) == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// 3️⃣ Передаем данные в службу нашего приложения.
	// Сервис возвращает структуру ExpandedURL
	newLink, err := h.service.ShortenURL(r.Context(), string(url)) // , userID

	// ✔️ Проверка на уникальность. iter13 ♊ пишет, что это правильно!
	// Сама проверка в методе Save (при записи в хранилище).
	// Здесь генерируем нужный ответ - 409
	var notUniqueErr *err_repo.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		// 4️⃣ Возвращаем клиенту response.
		publicationResult(w, h, newLink, http.StatusConflict)
		return
	}

	// Эта ошибка на тот случай, если все наши if не сработали.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicationResult(w, h, newLink, http.StatusCreated)
}
