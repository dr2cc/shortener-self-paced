package mw

import (
	"compress/gzip"
	"net/http"
)

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Если клиент прислал сжатые данные
		if r.Header.Get("Content-Encoding") == "gzip" {
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				http.Error(w, "invalid gzip body", http.StatusBadRequest)
				return
			}
			// defer gz.Close()
			// // Подменяем тело запроса на распакованный поток
			// r.Body = gz

			// Теперь мы НЕ закрываем gz здесь через defer,
			// а подменяем r.Body, чтобы Decode мог его прочитать.
			r.Body = gz
			// Важно: chi сам закроет r.Body после завершения запроса,
			// поэтому подмена на gzip.Reader безопасна.
		}

		next.ServeHTTP(w, r)
	})
}
