package mw

import (
	"compress/gzip"
	"net/http"
)

func DecompressRequest(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Проверяем заголовок сжатия
		if r.Header.Get("Content-Encoding") == "gzip" {
			// Создаем ридер. Если тело пустое, NewReader может вернуть EOF сразу.
			gz, err := gzip.NewReader(r.Body)
			if err != nil {
				// if err == io.EOF {
				// 	// Если тело просто пустое, идем дальше с оригинальным Body
				// 	next.ServeHTTP(w, r)
				// 	return
				// }

				// Если данные битые — возвращаем 400
				http.Error(w, err.Error(), http.StatusBadRequest)
				return
			}

			// // Оборачиваем закрытие: когда r.Body закроется, закроется и gzip
			// defer gz.Close()

			// Удаляем заголовок, чтобы хендлеры не пытались распаковать тело снова
			r.Header.Del("Content-Encoding")
			r.Header.Del("Content-Length") // Длина изменилась после распаковки

			r.Body = gz
		}

		next.ServeHTTP(w, r)
	})
}
