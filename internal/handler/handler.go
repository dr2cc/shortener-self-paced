// Package http_handlers contains functions that handles http requests
package handler

import (
	"app/internal/domain/link"
	"app/internal/lib/httpio"
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

// Интерфейс ShortenerUseCase ИСПОЛЬЗУЕТСЯ в пакете handler (РЕАЛИЗУЕТСЯ он в service),
// так как описывает всё, что Controller хочет от бизнес-логики. Интерфейс— это граница взаимодействия.
// ShortenerUseCase это интерфейс ко всем методам структуры Service.
type Shortener interface {
	FormatShortURL(urlID string) string
	ShortenURL(ctx context.Context, url string) (link.ExpandedURL, error)
	ShortenBatch(ctx context.Context, batch []service.BatchInput) ([]link.ExpandedURL, error)
	FindURL(ctx context.Context, id string) (link.ExpandedURL, error)
}

type PingerUseCase interface {
	CheckHealth(ctx context.Context) error
}

// Controller **контролирует** процесс обработки входящего запроса.
// Он не знает как работает бизнес-логика, но знает:
// 1️⃣ Как извлечь данные из HTTP-запроса (JSON, query-параметры).
// Как мапить❔❔
// 3️⃣ Какой метод сервиса вызвать.
// 4️⃣ Какой HTTP-статус и формат ответа вернуть клиенту.
// Controller в качестве методов имеет все эндпойнты и "инициализатор" роутера.
// Controller в качестве зависимости имеет указатель на структуру сервисов.
// Так обработчики передают свои запросы на уровень ниже- в слой сервисов❗
type Controller struct {
	//service *service.Service // Без интерфейсов. В таком случае интерфейсы здесь вообще не нужны
	shortener Shortener // Теперь здесь интерфейсы
	health    PingerUseCase
}

// Called from the app, creates a new instance of the Controller
func NewHandler(services *service.Service) *Controller {
	//func NewHandler(s ShortenerService) *Controller {
	return &Controller{
		shortener: services,
		health:    services,
	}
}

// Called from app
// InitRoutes — **карта маршрутов** приложения, определяющая точки входа API.
// Выбор буквы **h** для ресивера (получателя метода) явно указывает на роль Handler.
// Даже если структура называется Controller, её поведение — это обработка HTTP-запросов.
// Плюс: Сразу понятно, что это слой доставки (API).
func (h *Controller) InitRoutes(log *slog.Logger) chi.Router {
	router := chi.NewRouter()
	// 1. Базовое:
	// Присваиваем каждому запросу уникальный ID
	router.Use(middleware.RequestID)
	// Паника не должна ронять сервер
	router.Use(middleware.Recoverer) // Разместили здесь, чтобы страховать всё остальное
	// 2. ♊Кладем наш slog в контекст (чтобы httpio мог его достать)
	router.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// Добавляем ID запроса в логгер, чтобы все логи этого запроса были связаны
			logger := log.With(slog.String("request_id", middleware.GetReqID(r.Context())))
			// // Go запрещает использовать обычные строки в качестве ключей контекста,
			// // потому что два разных пакета могут использовать ключ "logger", и один затрет другой.
			// ctx := context.WithValue(r.Context(), "logger", logger)
			ctx := context.WithValue(r.Context(), httpio.LoggerKey, logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	// 3. Распаковываем входящий Gzip (если пришел gzip, он распаковывается и подменяет r.Body)
	router.Use(mw.DecompressRequest)
	// 4. Логируем сам факт запроса (к примеру URL, метод, время выполнения, статус)
	router.Use(mwLogger.New(log))
	// 5. Если клиент хочет сжатый ответ, запись в w перехватывается и сжимается (gzip)
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
