// Пакет entity определяет основные сущности для бизнес-логики (служб),
// а если применимо, то и для сопоставления базы данных и объектов HTTP-ответа.
// Каждая группа логических сущностей будет находится в отдельном файле.
package entity

// TODO👀: Переименовать в ExpandedURL (ex. ShortURL) - главная сущность сокращателя ссылок, содеожащаяя всю информацию
// необходимую для работы с URL
type ShortURL struct {
	OriginalURL string `json:"url"` // "url" - по заданию на инкремент 7, так называется ключ в JSON запросе
	ID          string `json:"id"`  // уникальный идентификатор (alias) для OriginalURL
	// TODO: изменить название
	CorrelationID string `json:"correlation_id"` //используется для сопоставления исходного и сокращённого URL при пакетном сокращении
	// CreatedByID   string `json:"created_by"`
	// DeletedAt     time.Time `json:"correlation_id"`
}
