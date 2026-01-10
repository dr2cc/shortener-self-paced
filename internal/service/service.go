// The service package contains the core business logic of the application.
package service

import (
	"app/internal/config"
	"app/internal/entity"
	storage "app/internal/repository"
	"context"
)

// type Service struct {
// 	// Сервис сокращения URL (создает ID (shortURL) из url), со своим функционалом
// 	Random     IDGenerator
// 	repository storage.Repository
// 	config     *config.Config
// }

// // Вызывается из app
// func NewService(rand IDGenerator, repo storage.Repository, conf *config.Config) *Service {
// 	return &Service{
// 		Random:     rand,
// 		repository: repo,
// 		config:     conf,
// 	}
// }

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL.
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/

// Сервис сокращения URL
type ShortURL interface {
	// Функцонал (видимо только обслуживающий работу хендлеров?):

	// // ❌ GenerateIDfromString непосредственно к работе хендлеров не относятся
	// // Генерация ID из строки URL
	// GenerateIDfromString(url string) (string, error)

	// FormatShortURL непосредственно к работе хендлеров не относятся, но ему нужна информация из cfg
	// (единственному, к стати)
	// Форматирование ID в результирующую строку
	FormatShortURL(urlID string) string
	// Мапим URL из запроса в структуру entity.ExpandedURL
	Shorten(ctx context.Context, url string) (entity.ExpandedURL, error)
	// Мапим массив входящих данных в []entity.ExpandedURL
	ShortenBatch(ctx context.Context, batch []entity.ExpandedURL) ([]entity.ExpandedURL, error)
	// Проверяем корректность работы выбранного хранилища
	HealthCheck(ctx context.Context) error
	// Находит в хранилище полный URL-адрес по указанному идентификатору.
	// Возвращает заполненную структуру entity.ExpandedURL
	FindURL(ctx context.Context, id string) (entity.ExpandedURL, error)
}

type Service struct {
	// Сервис сокращения URL (создает ID (shortURL) из url), со своим функционалом
	ShortURL
}

// Вызывается из app
func NewService(repos *storage.Repository, cfg *config.Config) *Service {
	return &Service{
		ShortURL: NewShortService(repos.ShortURL, cfg),
	}
}
