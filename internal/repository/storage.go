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

// (ex. hortURL, ex. Repository)
type ShortURLRepository interface {
	// Контракт интерфейса: интерфейс ShortURLRepository "обещает", что метод может вернуть ошибку.
	// Сервис обязан уважать этот контракт!
	Save(ctx context.Context, shortURL link.ExpandedURL) error
	FindByID(ctx context.Context, id string) (link.ExpandedURL, error)
	SaveBatch(ctx context.Context, batch []link.ExpandedURL) error
}

// Опциональный интерфес - его реализует только pg
type Pinger interface {
	CheckHealth(ctx context.Context) error
}

// Содержит интерфейс ShortURLRepository
type Repository struct {
	// Сервис сокращения URL, со своим функционалом
	ShortURLRepository
}

// // ❌ 27.01.2026 Проверить и удалить
// // Called from app
// // NewRepository возвращает единый объект,
// // который легко прокидывать в конструктор service.NewService ("чистая сигнатура")
// func NewRepository(cfg *config.Config, log *slog.Logger) *Repository {
// 	return &Repository{
// 		ShortURLRepository: ChoosingStorage(cfg, log),
// 	}
// }

// ♊ Called from app
func New(repo ShortURLRepository) *Repository {
	return &Repository{
		ShortURLRepository: repo,
	}
}

func ChoosingStorage(cfg *config.Config, log *slog.Logger) ShortURLRepository {
	if cfg.DatabaseDSN != "" {
		repo, err := pg.NewPostgresRepo(cfg, log)
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
