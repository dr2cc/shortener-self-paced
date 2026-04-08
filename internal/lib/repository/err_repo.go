package errrepo

import (
	"app/internal/domain/link"
	"errors"
)

// ♊
var (
	// ErrIDCollision сигнализирует о том, что сгенерированный короткий ID
	// уже существует в базе данных. Это техническая ошибка,
	// требующая повторной генерации ключа.
	ErrIDCollision = errors.New("short ID collision occurred")

	// ErrGenerationFailed используется, когда сервису не удалось
	// создать уникальный ID после максимально допустимого количества попыток.
	// Обычно указывает на исчерпание пространства имен или проблемы с рандомайзером.
	ErrGenerationFailed = errors.New("failed to generate unique short ID after maximum retries")
)

// NotUniqueURLError — ошибка, возникшая при сохранении URL, который уже существует.
type NotUniqueURLError struct {
	Err      error
	ShortURL link.ExpandedURL
}

func (err *NotUniqueURLError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueURLError) Unwrap() error {
	return err.Err
}

func NewNotUniqueURLError(shortURL link.ExpandedURL, err error) error {
	return &NotUniqueURLError{
		Err:      err,
		ShortURL: shortURL,
	}
}

// // Будем использовать для тестов
// var ErrNotUnique = func() error { return &NotUniqueURLError{} }()
