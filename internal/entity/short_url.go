// "Сущность" будет хранить структуру объекта и его метод (the entities will store the object’s struct and its method)
package entity

// ShortURL is main entity for system.
type ShortURL struct {
	// DeletedAt     time.Time
	OriginalURL string // исходный URL, который был сокращён
	ID          string // уникальный идентификатор короткого URL.
	// CreatedByID   string
	// CorrelationID string
}
