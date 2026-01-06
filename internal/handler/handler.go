// Package http_handlers contains functions that handles http requests
package handler

import (
	"app/internal/service"
	mwLogger "app/pkg/middleware/logger"
	"compress/flate"
	"log/slog"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ❗Логика работы слоя http обработчиков:
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) 1️⃣ данные по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.

// 01.01.2026 Как я теперь понимаю здесь (в handler) должны быть обязательно только три❗ сущности:
// 🔸Handler struct  - главное назначение- передача запросов на уровень ниже --> service
// 🔸func NewHandler - конструктор сущности Handler
// 🔸func InitRoutes - описание всех обработчиков
// Остальное- в зависимости от функционала приложения.
// Не нужно все сносить сюда! Распределять по слоям!

type Handler struct {
	service *service.Service // сервисы, содержащие бизнес-логику
}

// Вызывается ниже, из NewRouter
// func NewHandler(service *service.Service, config *config.Config) *Handler {
func NewHandler(service *service.Service) *Handler {
	return &Handler{
		service: service,
	}
}

// Вызывается из app
// InitRoutes создает новый маршрутизатор, добавляет middleware, а затем добавляет маршруты
func (h *Handler) InitRoutes(log *slog.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(flate.BestSpeed))

	// Получается service нужен только для работы ручек- передает в них
	// рандомайзер (бизнес-логику), хранилище и конфигурацию
	router.Get("/{id}", h.Redirect)
	router.Post("/", h.ShortenText)

	router.Post("/api/shorten", h.ShortenAPI)
	// iter10
	// Добавьте в сервис хендлер GET /ping,
	// который при запросе проверяет соединение с базой данных.
	// При успешной проверке хендлер должен вернуть HTTP-статус 200 OK, при неуспешной — 500 Internal Server Error.
	//
	router.Get("/ping", h.Ping)
	// iter12
	// Добавьте новый хендлер POST /api/shorten/batch,
	// принимающий в теле запроса множество URL для сокращения в формате:
	router.Post("/api/shorten/batch", h.BatchShortenAPI)

	return router
}

// func InitRoutes(service *service.Service, cfg *config.Config, log *slog.Logger) chi.Router {
// 	router := chi.NewRouter()

// 	router.Use(middleware.RequestID)
// 	router.Use(middleware.Logger)
// 	router.Use(mwLogger.New(log))
// 	router.Use(middleware.Recoverer)
// 	router.Use(middleware.Compress(flate.BestSpeed))

// 	// ❌ Переделать вызов сервера!! Как в todo-app1
// 	h := NewHandler(service, cfg)

// 	// Получается service нужен только для работы ручек- передает в них
// 	// рандомайзер (бизнес-логику), хранилище и конфигурацию
// 	router.Get("/{id}", h.Redirect)
// 	router.Post("/", h.ShortenText)

// 	router.Post("/api/shorten", h.ShortenAPI)
// 	// iter10
// 	// Добавьте в сервис хендлер GET /ping,
// 	// который при запросе проверяет соединение с базой данных.
// 	// При успешной проверке хендлер должен вернуть HTTP-статус 200 OK, при неуспешной — 500 Internal Server Error.
// 	//
// 	router.Get("/ping", h.Ping)
// 	// iter12
// 	// Добавьте новый хендлер POST /api/shorten/batch,
// 	// принимающий в теле запроса множество URL для сокращения в формате:
// 	router.Post("/api/shorten/batch", h.BatchShortenAPI)

// 	return router
// }
