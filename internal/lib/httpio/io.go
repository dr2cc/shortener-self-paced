package httpio

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Decode считывает, декомпрессирует и парсит JSON из запроса.
// Возвращает true, если всё успешно. При ошибке сама отправляет ответ клиенту.
func Decode(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	// Ограничиваем чтение (защита от слишком больших запросов)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	reader, err := getDecompressedReader(r)
	if err != nil {
		log.Printf("httpio: decompression error: %v", err)
		http.Error(w, "failed to decompress body", http.StatusInternalServerError)
		return false
	}
	defer reader.Close()

	if err := json.NewDecoder(reader).Decode(v); err != nil {
		log.Printf("httpio: decode error: %v", err)
		http.Error(w, "invalid json format", http.StatusBadRequest)
		return false
	}

	return true
}

// Respond сериализует данные и отправляет их с учетом поддержки gzip клиентом.
func Respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	var buf []byte
	var err error

	// 1. Сначала маршаллим в память, чтобы иметь возможность вернуть 500 при ошибке
	if data != nil {
		buf, err = json.Marshal(data)
		if err != nil {
			log.Printf("httpio: marshal error: %v", err)
			http.Error(w, "internal server error", http.StatusInternalServerError)
			return
		}
	}

	// 2. Устанавливаем базовый заголовок
	w.Header().Set("Content-Type", "application/json")

	// 3. Проверяем поддержку gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") && len(buf) > 0 {
		w.Header().Set("Content-Encoding", "gzip")
		w.WriteHeader(code)

		gz := gzip.NewWriter(w)
		if _, err := gz.Write(buf); err != nil {
			log.Printf("httpio: gzip write error: %v", err)
		}
		gz.Close()
		return
	}

	// 4. Обычный ответ без сжатия
	w.WriteHeader(code)
	if len(buf) > 0 {
		w.Write(buf)
	}
}

// Для BatchShortenAPI
// RespondStream используется для отправки больших объемов данных (слайсы, массивы).
// Данные кодируются сразу в сетевой поток, минуя создание большого буфера в памяти.
func RespondStream(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var writer io.Writer = w
	var gz *gzip.Writer

	// Проверяем поддержку gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		gz = gzip.NewWriter(w)
		writer = gz
		// В случае со стримом закрываем gzip в самом конце
		defer gz.Close()
	}

	// Отправляем статус ПЕРЕД кодированием
	w.WriteHeader(code)

	if data != nil {
		// Пишем JSON напрямую в поток (w или gz)
		if err := json.NewEncoder(writer).Encode(data); err != nil {
			// Здесь мы уже отправили Header, поэтому просто логируем ошибку.
			// Клиент получит оборванный JSON, что укажет ему на сбой.
			log.Printf("httpio: stream encoding error: %v", err)
		}
	}
}

// getDecompressedReader выбирает нужный ридер в зависимости от Content-Encoding
func getDecompressedReader(r *http.Request) (io.ReadCloser, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}
