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
	Save(ctx context.Context, shortURL entity.ShortURL) error
	FindByID(ctx context.Context, id string) (entity.ShortURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []entity.ShortURL) error
	// GetUsersUrls(ctx context.Context, userID string) ([]entity.ShortURL, error)
	// DeleteUrls(ctx context.Context, urls []entity.ShortURL) error
	// GetUsersAndUrlsCount(ctx context.Context) (int, int, error)
}

// NotUniqueURLError is error occurred when saving url is already exists.
type NotUniqueURLError struct {
	Err      error
	ShortURL entity.ShortURL
}

func (err *NotUniqueURLError) Error() string {
	return "url or id are already exist"
}

func (err *NotUniqueURLError) Unwrap() error {
	return err.Err
}

func NewNotUniqueURLError(shortURL entity.ShortURL, err error) error {
	return &NotUniqueURLError{
		Err:      err,
		ShortURL: shortURL,
	}
}

// // Будем использовать для тестов
// var ErrNotUnique = func() error { return &NotUniqueURLError{} }()
