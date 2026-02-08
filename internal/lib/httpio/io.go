package httpio

import (
	"encoding/json"
	"log/slog"
	"net/http"
)

// Вспомогательная функция для логгера. Безопасное извлечение логгера
func getLogger(r *http.Request) *slog.Logger {
	// // Go запрещает использовать обычные строки в качестве ключей контекста,
	// // потому что два разных пакета могут использовать ключ "logger", и один затрет другой.
	// if logger, ok := r.Context().Value("logger").(*slog.Logger); ok {
	if r != nil {
		if log, ok := r.Context().Value(LoggerKey).(*slog.Logger); ok && log != nil {
			return log
		}
	}
	return slog.Default()
}

// Decode теперь максимально простой
func Decode(w http.ResponseWriter, r *http.Request, v any) bool {
	log := getLogger(r)

	// Защита от слишком больших тел запроса
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024)

	// Просто декодируем r.Body.
	// Middleware уже распаковало его, если это был gzip.
	if err := json.NewDecoder(r.Body).Decode(v); err != nil {
		log.Debug("httpio: decode failed", slog.Any("err", err))
		Respond(w, r, http.StatusBadRequest, map[string]string{"error": "invalid json"})
		return false
	}

	return true
}

// Respond сериализует данные и отправляет их с учетом поддержки gzip клиентом.
// Respond теперь не думает о gzip — за это отвечает middleware.Compress
func Respond(w http.ResponseWriter, r *http.Request, code int, data any) {
	var buf []byte

	// 1. Сначала маршаллим в память, чтобы иметь возможность вернуть 500 при ошибке
	if data != nil {
		var err error
		buf, err = json.Marshal(data)
		if err != nil {
			getLogger(r).Error("marshal failed", slog.Any("err", err))
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	// 2. Устанавливаем базовый заголовок
	w.Header().Set("Content-Type", "application/json")

	// // 3. Безопасная проверка Gzip (r может быть nil)
	// canGzip := r != nil && strings.Contains(r.Header.Get("Accept-Encoding"), "gzip")

	// if canGzip && len(buf) > 0 {
	// 	w.Header().Set("Content-Encoding", "gzip")
	// 	w.WriteHeader(code)
	// 	gz := gzip.NewWriter(w)
	// 	_, _ = gz.Write(buf)
	// 	gz.Close()
	// 	return
	// }

	// 3. Обычный ответ
	w.WriteHeader(code)
	if len(buf) > 0 {
		w.Write(buf)
	}
}

// Для BatchShortenAPI
// RespondStream используется для отправки больших объемов данных (слайсы, массивы).
// Данные кодируются сразу в сетевой поток, минуя создание большого буфера в памяти.
func RespondStream(w http.ResponseWriter, r *http.Request, code int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(code)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		getLogger(r).Error("httpio: stream encode failed", slog.Any("err", err))
	}

	// var writer io.Writer = w
	// var gz *gzip.Writer

	// // Проверяем поддержку gzip
	// if strings.Contains(r.Header.Get("Accept-Encoding"), "gzip") {
	// 	w.Header().Set("Content-Encoding", "gzip")
	// 	gz = gzip.NewWriter(w)
	// 	writer = gz
	// 	// В случае со стримом закрываем gzip в самом конце
	// 	defer gz.Close()
	// }

	// // Отправляем статус ПЕРЕД кодированием
	// w.WriteHeader(code)

	// if data != nil {
	// 	// Пишем JSON напрямую в поток (w или gz)
	// 	if err := json.NewEncoder(writer).Encode(data); err != nil {
	// 		// Здесь мы уже отправили Header, поэтому просто логируем ошибку.
	// 		// Клиент получит оборванный JSON, что укажет ему на сбой.
	// 		log.Printf("httpio: stream encoding error: %v", err)
	// 	}
	// }
}
