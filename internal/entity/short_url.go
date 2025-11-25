// The entity package defines the core entities for business logic (services),
// а если применимо, то и для сопоставления базы данных и объектов HTTP-ответа.
// Каждая группа логических сущностей будет находится в отдельном файле.
package entity

// ExpandedURL (ex. ShortURL) - главная сущность сокращателя ссылок, содеожащаяя всю информацию
// необходимую для работы с URL
type ExpandedURL struct {
	OriginalURL   string `json:"url"`            // "url" - по заданию на инкремент 7, так называется ключ в JSON запросе
	ID            string `json:"id"`             // уникальный идентификатор (alias) для OriginalURL
	CorrelationID string `json:"correlation_id"` //используется для сопоставления исходного и сокращённого URL при пакетном сокращении
}
