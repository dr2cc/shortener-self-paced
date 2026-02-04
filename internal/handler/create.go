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

func (h *Controller) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var input dto.RequestShorten
	var notUniqueErr *err_repo.NotUniqueURLError

	// 1️⃣ Принимаем данные из сети, 2️⃣ десериализуем и заполняем (, &input) DTO
	if !httpio.Decode(w, r, &input) {
		return // Хелпер обработал все ошибки за нас, просто выходим
	}

	// Проверяем заполненность URL
	if input.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// Превращаем input (dto.RequestShorten) в serviceInput (service.SingleInput)
	serviceInput := service.SingleInput{
		URL: input.URL,
	}

	// TODO: явно корявая реализация ifs ❌ исправить!
	// 3️⃣ Отдаем в сервис "чистые" (без json из dto) данные и получаем данные для ответа
	link, err := h.service.ShortenURL(r.Context(), serviceInput.URL)
	if err == nil || errors.As(err, &notUniqueErr) {
		// Заполняем структуру для ответа
		output := dto.ResponseShorten{
			Result: h.service.FormatShortURL(link.ID),
		}
		if errors.As(err, &notUniqueErr) {
			// 4️⃣ Возвращаем клиенту response
			httpio.Respond(w, r, 409, output) // url не уникальный. iter13
			return
		}
		// 4️⃣ Возвращаем клиенту response
		httpio.Respond(w, r, 201, output) // успех
	}

	// В случае если ошибка не связана с уникальностью
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
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
