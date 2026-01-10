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
	// Сервис сокращения URL, со своим функционалом
	ShortURL
}

func NewRepository(log *slog.Logger, cfg *config.Config) *Repository {
	return &Repository{
		ShortURL: choosingStorage(log, cfg),
	}
}

func choosingStorage(log *slog.Logger, cfg *config.Config) ShortURL {
	if cfg.DatabaseDSN != "" {
		repo, err := pg.NewPostgresRepo(log, cfg)
		if err != nil {
			log.Error("failed to connect pg storage")
			os.Exit(1)
		}
		return repo
	}
	// Ментор считает, что это не отдельный вид хранилища,
	// а условие, что если есть env или флаг,
	// то надо попробовать прочитать файл, а по кончании в такой записать.
	// Судя по условию на Спринт 3, инкремент 11 ментор не прав:
	// "При отсутствии переменной окружения DATABASE_DSN или флага командной строки -d
	// или при их пустых значениях вернитесь последовательно к:
	// хранению сокращённых URL в файле при наличии соответствующей переменной окружения или флага командной строки;
	// хранению сокращённых URL в памяти."
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
