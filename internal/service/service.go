// The service package contains the core business logic of the application.
package service

import (
	"app/internal/config"
	"app/internal/domain/link"
	"app/internal/generator"
	storage "app/internal/repository"
	"context"
)

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL
// 🔸 проверка работоспособности db
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/

// Сервис сокращения URL
type ShortURL interface {
	// Функцонал:
	// Форматирование ID в результирующую строку, нужна информация из cfg (единственному!)
	FormatShortURL(urlID string) string
	// Мапим URL из запроса в структуру link.ExpandedURL
	CreateShortURL(ctx context.Context, url string) (link.ExpandedURL, error)
	// Мапим массив входящих данных в []link.ExpandedURL
	ShortenBatch(ctx context.Context, batch []link.ExpandedURL) ([]link.ExpandedURL, error)
	// Находит в хранилище полный URL-адрес по указанному идентификатору.
	// Возвращает заполненную структуру link.ExpandedURL
	FindURL(ctx context.Context, id string) (link.ExpandedURL, error)
}

// Сервис проверки работоспособности db
// Опциональный интерфес - его реализует только pg
type DBHealthChecker interface {
	// Функцонал:
	// Проверяем работоспособность db
	CheckHealth(ctx context.Context) error
}

type Service struct {
	// Сервис сокращения URL, со своим функционалом
	ShortURL
	// Сервис проверки работоспособности db, со своим функционалом
	DBHealthChecker
}

// Called from app
func NewService(repos *storage.Repository, gen *generator.StringGenerator, cfg *config.Config) *Service {
	return &Service{
		ShortURL:        NewShortService(repos.ShortURLRepository, gen, cfg),
		DBHealthChecker: NewHelthService(repos.ShortURLRepository),
	}
}
