// Package storage contains logic for saving and retrieving data.
package storage

import (
	"app/internal/config"
	"app/internal/entity"
	"app/internal/repository/pg"
	"context"
	"log/slog"
	"os"
)

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

var ErrNotUnique = func() error { return &NotUniqueURLError{} }()

// Repository saves and retrieves data from storage.
type Repository interface {
	Save(ctx context.Context, shortURL entity.ShortURL) error
	GetByID(ctx context.Context, id string) (entity.ShortURL, error)
	// GetUsersUrls(ctx context.Context, userID string) ([]entity.ShortURL, error)
	// Close(_ context.Context) error
	// Check(ctx context.Context) error
	// SaveBatch(ctx context.Context, batch []entity.ShortURL) error
	// DeleteUrls(ctx context.Context, urls []entity.ShortURL) error
	// GetUsersAndUrlsCount(ctx context.Context) (int, int, error)
}

// GetRepo is fabric that returns
// repository implementation based on cfg.
func GetRepo(log *slog.Logger, cfg *config.Config) Repository {
	if cfg.DatabaseDSN != "" {
		// repo, err := NewPgRepository(cfg.DatabaseDSN, cfg.MigrationsPath)
		// if err != nil {
		// 	panic(err)
		// }
		repo, err := pg.NewPostgresRepo(log, cfg)
		if err != nil {
			log.Error("failed to connect storage")
			os.Exit(1)
		}
		return repo
	}
	if cfg.FilePath != "" {
		repo, err := NewFileRepository(cfg.FilePath)
		if err != nil {
			panic(err)
		}
		return repo
	}

	return NewInMemoryRepository()
}
