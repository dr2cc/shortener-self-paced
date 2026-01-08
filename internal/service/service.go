// The service package contains the core business logic of the application.
package service

import (
	"app/internal/config"
	storage "app/internal/repository"
)

type Service struct {
	// Сервис сокращения URL (создает ID (shortURL) из url), со своим функционалом
	Random     IDGenerator
	repository storage.Repository
	config     *config.Config
}

// Вызывается из app
func NewService(rand IDGenerator, repo storage.Repository, conf *config.Config) *Service {
	return &Service{
		Random:     rand,
		repository: repo,
		config:     conf,
	}
}

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL.
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/
