package handler

import (
	"app/internal/domain/link"
	err_repo "app/internal/errors/repository"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// "result" - так в ответе, по заданию  на iter7, называется ключ в JSON запросе
type ResponseAPI struct {
	Result string `json:"result"`
}

func apiPublicationResult(w http.ResponseWriter, h *Controller, expandedURL link.ExpandedURL, status int) {
	res := ResponseAPI{
		Result: h.service.FormatShortURL(expandedURL.ID),
	}

	// marshalling - сортировка (сериализация) в json
	out, err := json.Marshal(res)
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

// 🤷‍♂️
// DTO. Локальная структура только для парсинга запроса
// Чтобы хендлер не "пачкал" Entity своими JSON-тегами,
// создал DTO прямо в пакете хендлера.
// Это позволит API меняться, не трогая бизнес-логику.
type shortenRequest struct {
	URL string `json:"url"`
}

// Назову json post ручку ShortenAPI - обычно при помощи json создают интерфейс
func (h *Controller) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var v shortenRequest

	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// десериализуем данные json
	if errDecode := json.NewDecoder(reader).Decode(&v); errDecode != nil {
		http.Error(w, "cannot decode json", http.StatusBadRequest)
		return
	}

	// Так как анмарщаллинг происходит с данными в ExpandedURL,
	// то структурный тег поля OriginalURL в ExpandedURL ("url")
	// должен совпадать с ключем "url" запроса, иначе будет "url required"
	if v.URL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	// Если нет ошибок и значение url уникально, то err == nil
	// Если url не уникален, то
	// err == entity.ExpandedURL{OriginalURL: url, ID:urlID},  &shorteningError{Err:err, ShortURL: shortURL}
	expandedURL, err := h.service.CreateShortURL(r.Context(), v.URL)

	// Iter13 генерация нужного ответа - 409
	var notUniqueErr *err_repo.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		apiPublicationResult(w, h, expandedURL, http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiPublicationResult(w, h, expandedURL, http.StatusCreated)
}

func publicationResult(w http.ResponseWriter, h *Controller, expandedURL link.ExpandedURL, status int) {
	w.Header().Set("Content-Type", "text/html")
	w.WriteHeader(status)
	content := h.service.FormatShortURL(expandedURL.ID)
	if _, err := w.Write([]byte(content)); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// ❌ 18.01.26 не пойму как мой ShortenText выполняет структурирование (маппинг) в ExpandedURL ??
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
	expandedURL, err := h.service.CreateShortURL(r.Context(), string(url)) // , userID

	// ✔️ Проверка на уникальность. iter13 ♊ пишет, что это правильно!
	// Сама проверка в методе Save (при записи в хранилище).
	// Здесь генерируем нужный ответ - 409
	var notUniqueErr *err_repo.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		// 4️⃣ Возвращаем клиенту response.
		publicationResult(w, h, expandedURL, http.StatusConflict)
		return
	}

	// Эта ошибка на тот случай, если все наши if не сработали.
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicationResult(w, h, expandedURL, http.StatusCreated)
}
