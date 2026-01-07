// The service package contains the core business logic of the application.
package service

import (
	"app/internal/config"
	storage "app/internal/repository"
)

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL.
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/
type Service struct {
	// Сервис сокращения URL (создает ID (shortURL) из url), со своим функционалом
	Random IDGenerator
	// ❌ по todo-app, "общаться" с хранилищем правильнее из самого сервиса (пока только Random)
	repository storage.Repository
	config     *config.Config
	//generator  generator.URLGenerator
}

// Вызывается из app
func NewService(rand IDGenerator, repo storage.Repository, conf *config.Config) *Service {
	return &Service{
		Random: rand,
		// // ❌ по todo-app вот так создают сервисы приложения
		// // Создаю новый проект. Назову shortener-todo-app. В нем iter1 и затем iter13(!?)
		// // В iter1 привожу к todo-app (м.б. и gin использовать?!)
		// Random:     NewRandomService(repo.Random),
		repository: repo,
		config:     conf,
		//generator:  generator,
	}
}
