// Package storage contains logic for saving and retrieving data.
package storage

import (
	"app/internal/domain/link"
	"context"
)

// Интерфейс это Контракт:
// ShortURLRepository "обещает", что метод может вернуть ошибку.
// Сервис обязан уважать этот контракт!
type ShortURLRepository interface {
	Save(ctx context.Context, shortURL link.ExpandedURL) error
	FindByID(ctx context.Context, id string) (link.ExpandedURL, error)
	SaveBatch(ctx context.Context, batch []link.ExpandedURL) error
}

// // Опциональный интерфес - его реализует только pg
// type Pinger interface {
// 	CheckHealth(ctx context.Context) error
// }

// Содержит интерфейс ShortURLRepository
type Repository struct {
	// Сервис сокращения URL, со своим функционалом
	ShortURLRepository
}

// ♊ Called from app
func New(repo ShortURLRepository) *Repository {
	return &Repository{
		ShortURLRepository: repo,
	}
}
