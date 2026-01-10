package errrepo

import "app/internal/entity"

// NotUniqueURLError — ошибка, возникшая при сохранении URL, который уже существует.
type NotUniqueURLError struct {
	Err      error
	ShortURL entity.ExpandedURL
}

func (err *NotUniqueURLError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueURLError) Unwrap() error {
	return err.Err
}

func NewNotUniqueURLError(shortURL entity.ExpandedURL, err error) error {
	return &NotUniqueURLError{
		Err:      err,
		ShortURL: shortURL,
	}
}

// // Будем использовать для тестов
// var ErrNotUnique = func() error { return &NotUniqueURLError{} }()
