package handlers

import (
	"app/internal/entity"
	"app/internal/storage"
	"encoding/json"
	"errors"
	"io"
	"net/http"
)

// "result" - так в ответе, по заданию  на инкремент 7, называется ключ в JSON запросе
type ResponseAPI struct {
	Result string `json:"result"`
}

func apiPublicationResult(w http.ResponseWriter, h *Handler, shortURL entity.ExpandedURL, status int) {
	res := ResponseAPI{
		Result: h.service.FormatShortURL(shortURL.ID),
	}

	// marshalling - сортировка (сериаоизация) в json
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

// Назову json post ручку ShortenAPI - обычно при помощи json создают интерфейс
func (h *Handler) ShortenAPI(w http.ResponseWriter, r *http.Request) {
	var v entity.ExpandedURL

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

	// Так как анмарщаллинг происходит с данными в entity.ExpandedURL (ex. ShortURL),
	// то структурный тег поля OriginalURL в entity.ExpandedURL ("url")
	// должен совпадать с ключем "url" запроса, иначе будет "url required"
	if v.OriginalURL == "" {
		http.Error(w, "url required", http.StatusBadRequest)
		return
	}

	//userID := h.getUserID(r)

	// Если нет ошибок и значение url уникально, то err == nil
	// Если url не уникален, то
	// err == entity.ExpandedURL{OriginalURL: url, ID:urlID},  &shorteningError{Err:err, ShortURL: shortURL}
	shortURL, err := h.service.Shorten(r.Context(), v.OriginalURL)

	// Iter13 генерация нужного ответа - 409
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		apiPublicationResult(w, h, shortURL, http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	apiPublicationResult(w, h, shortURL, http.StatusCreated)
}

func publicationResult(w http.ResponseWriter, h *Handler, shortURL entity.ExpandedURL, status int) {
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

	// Если нет ошибок и значение url уникально, то err == nil
	// Если url не уникален, то
	// err == entity.ExpandedURL{OriginalURL: url, ID:urlID},  &shorteningError{Err:err, ShortURL: shortURL}
	shortURL, err := h.service.Shorten(r.Context(), string(url)) // , userID

	// iter13. Проверка на уникальность
	// Сама проверка в Shorten, а точнеее в методе Save (при записи в хранилище).
	// Здесь генерируем нужный ответ - 409
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		publicationResult(w, h, shortURL, http.StatusConflict)
		return
	}

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	publicationResult(w, h, shortURL, http.StatusCreated)
}
