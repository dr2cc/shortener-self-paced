package handler

import (
	"app/internal/domain/link"
	"app/internal/lib/api/dto"
	err_repo "app/internal/lib/repository"
	"app/internal/service"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

func (h *Controller) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	// 1️⃣ Принимаем данные.
	var input dto.RequestShorten

	// TODO: Вынести анмаршалинг в отдельную функцию? Оформить отдельным пакетом?

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	if errDecode := json.NewDecoder(reader).Decode(&input); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	} // 2️⃣ Десериализуем (анмаршалинг) данные из сети и заполняем (.Decode(&input)) DTO

	// Проверяем заполненность URL
	if input.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// Превращаем input (dto.RequestShorten) в serviceInput (service.SingleInput)
	serviceInput := service.SingleInput{
		URL: input.URL,
	}

	// 3️⃣ Отдаем в сервис "чистые" (без json из dto) данные и получаем данные для ответа
	link, err := h.service.ShortenURL(r.Context(), serviceInput.URL)
	// Если нет ошибок и значение url уникально, то
	// err == nil
	// Если url не уникален, то
	// err == &shorteningError{Err:err, ShortURL: shortURL}
	var notUniqueErr *err_repo.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		// url не уникальный. iter13
		responseShortenAPI(w, h, link, http.StatusConflict)
		return
	}
	// В случае если ошибка не связана с уникальностью
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// успех
	responseShortenAPI(w, h, link, http.StatusCreated)
}

func responseShortenAPI(w http.ResponseWriter, h *Controller, expandedURL link.ExpandedURL, status int) {
	// Заполняем структуру для ответа
	output := dto.ResponseShorten{
		Result: h.service.FormatShortURL(expandedURL.ID),
	}

	// 4️⃣ Сериалиазуем (marshalling).
	resp, err := json.Marshal(output)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if _, err = w.Write(resp); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
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
	reader, err := getDecompressedReader(r)
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
