// "Сущность" будет хранить структуру объекта и его метод
// The entities will store the object’s struct and its method
package entity

// ShortURL главная сущность (entity) проекта
type ShortURL struct {
	OriginalURL string `json:"url"`
	ID          string `json:"id"` // уникальный идентификатор (alias) для OriginalURL
	// TODO: изменить название
	CorrelationID string `json:"created_by"` //используется для сопоставления исходного и сокращённого URL при пакетном сокращении
	// CreatedByID   string `json:"created_by"`
	// DeletedAt     time.Time `json:"correlation_id"`
}
