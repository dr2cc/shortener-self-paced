// DTO (Data Transfer Object) нужны исключительно как «переводчики» между внешним миром (JSON)
// и кодом проекта (Service/Domain).
// ❗В идеальной чистой архитектуре service не должен знать о пакете dto.
package dto

import "errors"

// Делаем структуру dto.RequestShortenBatch «умной».
var ErrEmptyURL = errors.New("url required in batch item")

type RequestShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

// BatchRequest — это тип-обертка для слайса структр RequestShortenBatch,
// чтобы добавить ему поведение (Validate)
type BatchRequest []RequestShortenBatch

// Validate проверяет весь массив данных разом
func (b BatchRequest) Validate() error {
	for _, item := range b {
		if item.OriginalURL == "" {
			return ErrEmptyURL
		}
	}
	return nil
}

// CreateBatchResponse (ex. ShorteningBatchResult) — результат сокращения при пакетном вводе
type ResponseShortenBatch struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}

// DTO. Cтруктура только для парсинга запроса.
// Чтобы хендлер не "пачкал" Entity своими JSON-тегами,
// создал DTO.
// Это позволит API меняться, не трогая бизнес-логику.
type RequestShorten struct {
	URL string `json:"url"`
}

// "result" - так в ответе, по заданию  на iter7, называется ключ в JSON запросе
type ResponseShorten struct {
	Result string `json:"result"`
}
