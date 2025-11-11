// "Сущность" будет хранить структуру объекта и его метод (the entities will store the object’s struct and its method)
package entity

// ShortURL главная сущность (entity) проекта
type ShortURL struct {
	OriginalURL string
	ID          string // уникальный идентификатор (alias) для OriginalURL
	// CreatedByID   string
	// CorrelationID string
	// DeletedAt     time.Time
}
