package link

// Entity — сердце системы.
// Здесь НЕТ тегов json, потому что домену плевать на HTTP.
// ♊ советует переименовать в Link
// ❌ переименую после контроля за использованием, особенно в слое handler
type ExpandedURL struct {
	OriginalURL   string
	ID            string
	CorrelationID string
}
