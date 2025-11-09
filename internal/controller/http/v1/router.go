package v1

import (
	"app/internal/config"
	"app/internal/controller/http/v1/save"
	"app/internal/controller/ping"
	"app/internal/repository/pg"
	myLog "app/internal/usecase/middleware/logger"
	"app/internal/usecase/random"
	"log/slog"

	"github.com/go-chi/chi"
	"github.com/go-chi/chi/middleware"
)

func Router(router *chi.Mux, cfg *config.Config, repoPg *pg.PostgresRepo, randomKey random.RandomGenerator, log *slog.Logger) {
	// Middleware встроенный в chi
	router.Use(middleware.RequestID) // Трассировка. Добавляется request_id в каждый запрос
	router.Use(middleware.Logger)    // Логирование всех запросов
	// Если внутри произойдет паника, приложение не упадет.
	// Recoverer это compress.Gzipper, которое восстанавливается после паники,
	// регистрирует панику и выводит идентификатор запроса, если он указан.
	router.Use(middleware.Recoverer)
	router.Use(myLog.New(log))       // Меняю логгер на мой
	router.Use(middleware.URLFormat) // Парсер URLов поступающих запросов. Удалит суффикс из пути маршрутизации и продолжит маршрутизацию

	handlers := save.Handler{
		Repo: repoPg,
	}

	// handlers
	router.Get("/healthDB", ping.HealthCheck(repoPg, log))
	//router.Post("/", save.New(repo, randomKey, log))
	router.Post("/", handlers.New(randomKey, log))

}
