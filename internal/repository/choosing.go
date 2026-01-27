package storage

import (
	"app/internal/config"
	"app/internal/repository/cache"
	jsonstore "app/internal/repository/jsonrstore"
	"app/internal/repository/pg"
	"log/slog"
	"os"
)

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
