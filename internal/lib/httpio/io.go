package httpio

import (
	"compress/gzip"
	"encoding/json"
	"io"
	"log"
	"net/http"
	"strings"
)

// Decode скрывает всю кухню с декомпрессией и парсингом
func Decode(w http.ResponseWriter, r *http.Request, v interface{}) bool {
	// 1. Ограничиваем размер (защита от переполнения памяти)
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	// 2. Выбираем правильный ридер (обычный или gzip)
	reader, err := getDecompressedReader(r)
	if err != nil {
		http.Error(w, "Decompression failed: "+err.Error(), http.StatusInternalServerError)
		return false
	}
	defer reader.Close()

	// 3. Декодируем JSON
	if err := json.NewDecoder(reader).Decode(v); err != nil {
		http.Error(w, "Cannot decode JSON", http.StatusBadRequest)
		return false
	}

	return true
}

// getDecompressedReader — та самая логика, вынесенная в хелпер
func getDecompressedReader(r *http.Request) (io.ReadCloser, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

// // Respond для быстрой отправки ответов
// func Respond(w http.ResponseWriter, code int, data interface{}) {
// 	w.Header().Set("Content-Type", "application/json")
// 	w.WriteHeader(code)
// 	if data != nil {
// 		_ = json.NewEncoder(w).Encode(data)
// 	}
// }

// Respond — отправляет JSON, автоматически сжимая его при необходимости
func Respond(w http.ResponseWriter, r *http.Request, code int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")

	var writer io.Writer = w

	// Проверяем, поддерживает ли клиент gzip
	if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
		w.Header().Set("Content-Encoding", "gzip")
		// Создаем gzip writer, который будет писать в исходный ResponseWriter
		gz := gzip.NewWriter(w)
		// Важно: закрываем gzip writer в конце функции
		// Это гарантирует, что все сжатые данные будут записаны в ResponseWriter
		defer gz.Close()
		writer = gz // Теперь запись идет через gzip
	}

	// Устанавливаем статус-код
	w.WriteHeader(code)

	// Кодируем данные в JSON и записываем в writer (либо gzip, либо обычный)
	if data != nil {
		if err := json.NewEncoder(writer).Encode(data); err != nil {
			// Если произошла ошибка ПРИ записи ответа:
			// 1. Мы не можем отправить http.Error, так как статус-код уже отправлен.
			// 2. Единственный вариант — залогировать ошибку для администратора.
			log.Printf("ERROR: httpio.Respond failed to encode/write response: %v", err)
			// TODO: передавать сюда логгер
		}
	}
}
