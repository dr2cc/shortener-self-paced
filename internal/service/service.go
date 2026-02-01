// The service package contains the core business logic of the application.
package service

import (
	"app/internal/domain/link"
	"context"
)

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL
// 🔸 проверка работоспособности db
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/

// Сервис сокращения URL - "Что мы делаем?" (интерфейс).
// Если ShortURL в сервисе будет повторять методы ShortURLRepository (здесь не так) из репозитория
//
//	— это правильно (принцип инверсии зависимостей).
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
// Что мы делаем (интерфейс)
type Pinger interface {
	// Функцонал:
	// Проверяем работоспособность db
	CheckHealth(ctx context.Context) error
}

// noOpPinger — "заглушка-оптимист" для хранилищ без поддержки Ping (In-mem, File).
type noOpPinger struct{}

// По "подсказкам" все CheckHealth должны быть Ping.
// Вообще не важно! Главное не запутаться.
func (n *noOpPinger) CheckHealth(ctx context.Context) error {
	return nil // Всегда "здоров"
}

// Service — **единая точка входа** (агрегатор или структурная обёртка)
// в бизнес-логику и средства контроля инфраструктуры (тут db).
// Как мы делаем (структура и логика)
type Service struct {
	// Сервис сокращения URL, со своим функционалом
	ShortURL
	// Сервис проверки работоспособности db, со своим функционалом
	Pinger
}

// // Called from app
// func NewService(repos *storage.Repository, gen *generator.StringGenerator, cfg *config.Config) *Service {
// 	return &Service{
// 		ShortURL: NewShortService(repos.ShortURLRepository, gen, cfg),
// 		Pinger:   NewHelthService(repos.ShortURLRepository),
// 	}
// }

// ♊ Конструктор агрегатора
func New(links ShortURL, repo any) *Service {
	svc := &Service{ShortURL: links}

	// Проверяем: если репозиторий поддерживает Ping, используем его
	if p, ok := repo.(Pinger); ok {
		svc.Pinger = p
	} else {
		// Если это файл или память — ставим "заглушку-оптимист"
		svc.Pinger = &noOpPinger{}
	}
	return svc
}
