// Package http_handlers contains functions that handles http requests
package handler

import (
	mw "app/internal/lib/middleware"
	"app/internal/service"
	mwLogger "app/pkg/middleware/logger"
	"compress/flate"
	"context"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

// ❗Логика работы слоя http обработчиков:
// 1️⃣ Принимаем данные от клиента (обычно в формате json).
// 2️⃣ Мапим (преобразуем в конкретную объектную модель, структуру) данные 1️⃣ по нашей внутренней структуре.
// 3️⃣ Передаем данные в службу нашего приложения.
// 4️⃣ Возвращаем клиенту response.
// Еще этот слой называют ❗Delivery.
// Суть: Это точка входа в приложение для внешнего мира.
// Этот слой «доставляет» данные из внешнего протокола (HTTP, gRPC, CLI) внутрь бизнес-логики и обратно.

// Controller **контролирует** процесс обработки входящего запроса.
// Он не знает как работает бизнес-логика, но знает:
// 1️⃣ Как извлечь данные из HTTP-запроса (JSON, query-параметры).
// 3️⃣ Какой метод сервиса вызвать.
// 4️⃣ Какой HTTP-статус и формат ответа вернуть клиенту.
// Controller в качестве методов имеет все эндпойнты и "инициализатор" роутера.
// Controller в качестве зависимости имеет указатель на структуру сервисов.
// Так обработчики передают свои запросы на уровень ниже- в слой сервисов❗
type Controller struct {
	service *service.Service // сервисы, содержащие бизнес-логику
}

// Called from the app, creates a new instance of the Controller
func NewHandler(service *service.Service) *Controller {
	return &Controller{
		service: service,
	}
}

// Called from app
// InitRoutes — **карта маршрутов** приложения, определяющая точки входа API.
// Выбор буквы **h** для ресивера (получателя метода) явно указывает на роль Handler.
// Даже если структура называется Controller, её поведение — это обработка HTTP-запросов.
// Плюс: Сразу понятно, что это слой доставки (API).
func (h *Controller) InitRoutes(log *slog.Logger) chi.Router {
	router := chi.NewRouter()

	// 1. Присваиваем каждому запросу уникальный ID
	router.Use(middleware.RequestID)

	// router.Use(middleware.Logger)
	// 2. ♊Кладем наш slog в контекст (чтобы httpio мог его достать)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем ID запроса в логгер, чтобы все логи этого запроса были связаны
			logger := log.With(slog.String("request_id", middleware.GetReqID(r.Context())))
			ctx := context.WithValue(r.Context(), "logger", logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})

	// 3. Логируем сам факт запроса (к примеру URL, метод, время выполнения, статус)
	router.Use(mwLogger.New(log))
	// 4. Распаковываем входящий Gzip (если пришел gzip,
	// он распаковывается и подменяет r.Body)
	router.Use(mw.DecompressRequest)
	// 5. Паника не должна ронять сервер
	router.Use(middleware.Recoverer)
	// 6. Если клиент хочет сжатый ответ,
	// запись в w перехватывается и сжимается (gzip)
	router.Use(middleware.Compress(flate.BestSpeed))

	// Роуты
	// Service Pinger
	router.Get("/ping", h.Ping) // iter10
	// Service ShortURL
	router.Post("/", h.ShortenText)
	router.Get("/{id}", h.Redirect)
	router.Post("/api/shorten", h.ShortenAPI)            // iter7
	router.Post("/api/shorten/batch", h.BatchShortenAPI) // iter12

	return router
}
