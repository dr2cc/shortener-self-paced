// Service Input (структуры-посредники = собственные примитивы)
package service

// Чтобы сервис стал по-настоящему независимым («чистым»),
// он должен принимать данные в своих собственных терминах или в примитивах.

// ❌ Вероятно тоже не нужен?
type BatchInput struct {
	CorrelationID string
	OriginalURL   string
}
