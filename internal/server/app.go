package server

import (
	"app/internal/config"
	v1 "app/internal/controller/http/v1"
	"app/internal/repository/pg"
	"app/internal/usecase/logger/sl"
	"app/internal/usecase/random"
	"app/pkg/httpserver"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

// - 2️⃣ Внедрение зависимостей.
// 		Инициализируйте и соедините (это и есть внедрение зависимостей)
// 		основные компоненты вашего приложения, такие как:
//  -- клиент базы данных,
//  -- уровень хранения (это компонент приложения, отвечающий за абстрагирование
// и управление взаимодействием с источником данных)
//  -- обработчики запросов (здесь они Application or Busines Logic).

// - 3️⃣ Настройка маршрутизатора: создайте экземпляр маршрутизатора Chi и зарегистрируйте маршруты,
// передав обработчики из логического уровня вашего приложения.

// - 4️⃣ Запуск сервера: запустите HTTP-сервер, обычно с помощью http.ListenAndServe, и корректно обработайте возможные ошибки запуска.

// В целом не до конца понимаю, что это дает (04.11.2025)
// но давно хотел создать в конструкторе "главную" структуру приложения
type App struct {
	httpServer *http.Server
}

func NewApp() *App {
	return &App{}
}

// Run creates objects (via constructors!)
func (a *App) Run(cfg *config.Config) {
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.HTTPServer.Address)) // Помимо сообщения выведем параметр с адресом
	log.Debug("logger debug mode enabled")

	// Repository🧹🏦
	repo, err := pg.NewPostgresRepo(log, cfg)
	if err != nil {
		log.Error("failed to connect storage")
		os.Exit(1)
	}

	// Use-Case🧹🏦
	// В данный момент именно service я не создаю. Сложно..
	// Видимо им можно считать вызов server.NewApp в main
	randomKey := random.RandomGenerator{}

	// HTTP Server🧹🏦
	mux := chi.NewRouter()
	// middlewares & handlers
	v1.Router(mux, cfg, repo, randomKey, log)
	a.httpServer = httpserver.New(cfg.HTTPServer.Address, mux, log)

	// Waiting signal🧹🏦
	done := make(chan os.Signal, 1)
	signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	<-done
	log.Info("stopping server")

	// Смысл таймаута был, но сейчас потерян..
	ctx := context.Background() //context.WithTimeout(context.Background(), 10*time.Second)
	//defer cancel()

	// Shutdown🧹🏦
	if err := a.httpServer.Shutdown(ctx); err != nil {
		log.Error("failed to stop server", sl.Err(err))
		return
	}

	// TODO: close storage

	log.Info("server stopped")
}

func setupLogger(env string) *slog.Logger {
	var log *slog.Logger

	switch env {
	case envLocal:
		log = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envDev:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelDebug}))
	case envProd:
		log = slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: slog.LevelInfo}))
	}

	return log
}
