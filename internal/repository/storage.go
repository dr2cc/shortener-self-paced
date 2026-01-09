// Package storage contains logic for saving and retrieving data.
package storage

import (
	"app/internal/config"
	"app/internal/entity"
	"app/internal/repository/cache"
	jsonstore "app/internal/repository/jsonrstore"
	"app/internal/repository/pg"
	"context"
	"log/slog"
	"os"
)

// ex. Repository
type ShortURL interface {
	Save(ctx context.Context, shortURL entity.ExpandedURL) error
	FindByID(ctx context.Context, id string) (entity.ExpandedURL, error)
	Close(_ context.Context) error
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []entity.ExpandedURL) error
}

type Repository struct {
	ShortURL
}

func NewRepository(log *slog.Logger, cfg *config.Config) *Repository {
	return &Repository{
		ShortURL: choosingStorage(log, cfg),
	}
}

// type ChoosingStorage interface {
// 	NewPostgresRepo(log *slog.Logger, cfg *config.Config)
// 	NewFileRepository(cfg *config.Config)
// 	NewInMemoryRepository()
// }

func choosingStorage(log *slog.Logger, cfg *config.Config) ShortURL {
	if cfg.DatabaseDSN != "" {
		repo, err := pg.NewPostgresRepo(log, cfg)
		if err != nil {
			log.Error("failed to connect pg storage")
			os.Exit(1)
		}
		return repo
	}
	if cfg.FilePath != "" {
		repo, err := jsonstore.NewFileRepository(cfg.FilePath)
		if err != nil {
			log.Error("file (jsonstore) storage error")
			os.Exit(1)
		}
		return repo
	}

	return cache.NewInMemoryRepository()
}
