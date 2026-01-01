// Package http_handlers contains functions that handles http requests
package handlers

import (
	"app/internal/config"
	"app/internal/service"
	mwLogger "app/pkg/middleware/logger"
	"compress/flate"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// 01.01.2026 Как я теперь понимаю здесь должны быть обязательно только три❗ сущности:
// 🔸Handler struct  - главное назначение- передача запросов на уровень ниже --> service
// 🔸func NewHandler - конструктор сущности Handler
// 🔸func NewRouter (InitRoutes правильнее, это не конструктор) - описание всех обработчиков
// Остальное- в зависимости от функционала приложения.
// Не нужно все сносить сюда! Распределять по слоям!

type Handler struct {
	Mux     *chi.Mux         // маршрутизатор, который мы будем использовать для обработки запросов
	service *service.Service // сервисы, содержащие бизнес-логику
	// crypto
}

// Вызывается ниже, из NewRouter
// NewHandler создает новый экземпляр структуры Handler, инициализирует chi мультиплексор,
// и выбирает службу
// и crypto
func NewHandler(service *service.Service, config *config.Config) *Handler {
	// cryptographer := {}
	return &Handler{
		Mux:     chi.NewMux(),
		service: service,
		// crypto:  &cryptographer,
	}
}

// Вызывается из app
// NewRouter создает новый маршрутизатор, добавляет middleware, а затем добавляет маршруты
func NewRouter(service *service.Service, cfg *config.Config, log *slog.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, cfg)

	// Получается service нужен только для работы ручек- передает в них
	// рандомайзер (бизнес-логику), хранилище и конфигурацию
	router.Get("/{id}", h.Redirect)
	router.Post("/", h.ShortenText)
	// // При простой аутентификации, можно использовать такую конструкцию:
	// router.Route("/", func(r chi.Router) {
	// 	r.Use(middleware.BasicAuth("url-shortener", map[string]string{
	// 		cfg.User: cfg.Password,
	// 	}))
	// 	r.Post("/", h.ShortText)
	// })

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

	// //************************************************************************************
	// // iter14
	// // 	Добавьте в сервис функциональность аутентификации пользователя.
	// // Сервис должен:
	// // ◽ Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя,
	// // если такой куки не существует или она не проходит проверку подлинности.

	// // ◽ Иметь хендлер GET /api/user/urls,
	// // который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
	// // [
	// //     {
	// //         "short_url": "http://...",
	// //         "original_url": "http://..."
	// //     },
	// //     ...
	// // ]
	// router.Get("/api/user/urls", h.UserURLs)
	// //

	return router
}

// Если тело запроса сжато с помощью gzip, возвращает gzip reader,
// в противном случае возвращает request body (default reader)
func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

// Ping is a health check endpoint.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
