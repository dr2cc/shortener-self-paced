// Package storage contains logic for saving and retrieving data.
package storage

import (
	"app/internal/entity"
	"context"
)

// Интерфейс создает "границу" обмена информацией между разными частями проекта
// Интерфейс Repository определяет работу с хранением данных в проекте
// DataStorageMethods
type Repository interface {
	Save(ctx context.Context, shortURL entity.ExpandedURL) error
	FindByID(ctx context.Context, id string) (entity.ExpandedURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []entity.ExpandedURL) error
	GetUsersUrls(ctx context.Context, userID string) ([]entity.ExpandedURL, error)
}

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
