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
				// // Если данные битые — возвращаем 400
				// http.Error(w, "gzip: "+err.Error(), http.StatusBadRequest)
				// return
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

// func DecompressRequest(next http.Handler) http.Handler {
// 	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 		// Если клиент прислал сжатые данные
// 		if r.Header.Get("Content-Encoding") == "gzip" {
// 			gz, err := gzip.NewReader(r.Body)
// 			if err != nil {
// 				http.Error(w, "invalid gzip body", http.StatusBadRequest)
// 				return
// 			}
// 			// defer gz.Close()
// 			// // Подменяем тело запроса на распакованный поток
// 			// r.Body = gz

// 			// Теперь мы НЕ закрываем gz здесь через defer,
// 			// а подменяем r.Body, чтобы Decode мог его прочитать.
// 			r.Body = gz
// 			// Важно: chi сам закроет r.Body после завершения запроса,
// 			// поэтому подмена на gzip.Reader безопасна.
// 		}

// 		next.ServeHTTP(w, r)
// 	})
// }
