// Package storage contains logic for saving and retrieving data.
package storage

import (
	"app/internal/config"
	"app/internal/domain/link"
	"app/internal/repository/cache"
	jsonstore "app/internal/repository/jsonrstore"
	"app/internal/repository/pg"
	"context"
	"log/slog"
	"os"
)

// 🤷‍♂️ Переименовать в ShortURLRepository (ex. Repository)
type ShortURL interface {
	// Контракт интерфейса: интерфейс ShortURL "обещает", что метод может вернуть ошибку.
	// Сервис обязан уважать этот контракт!
	Save(ctx context.Context, shortURL link.ExpandedURL) error
	FindByID(ctx context.Context, id string) (link.ExpandedURL, error)
	// Close(_ context.Context) error //🤷‍♂️ нет в задании- не нужен
	Check(ctx context.Context) error
	SaveBatch(ctx context.Context, batch []link.ExpandedURL) error
}

// 🤷‍♂️ Лишнее?? Сравнить с Жашкевичем
type Repository struct {
	// Сервис сокращения URL, со своим функционалом
	ShortURL
}

func NewRepository(cfg *config.Config, log *slog.Logger) *Repository {
	return &Repository{
		ShortURL: choosingStorage(cfg, log),
	}
}

func choosingStorage(cfg *config.Config, log *slog.Logger) ShortURL {
	// ❌ 13.01.2026 Убрал остальные хранилища (до полного изменения кода)
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
	// 🤷‍♂️ Условие не верное! Точнее плохо сформулированное (на Спринт 3, инкремент 11):
	// "При отсутствии переменной окружения DATABASE_DSN или флага командной строки -d
	// или при их пустых значениях вернитесь последовательно к:
	// 🤷‍♂️❗ Один раз!! Не хранилище!
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
