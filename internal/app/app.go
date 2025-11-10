package app

import (
	"app/internal/config"
	storage "app/internal/repository"
	"app/internal/server"
	"app/internal/usecase/logger/sl"
	"app/internal/usecase/random"
	services "app/internal/usecase/shortener"
	"context"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

type App struct {
	httpServer *http.Server
}

func NewApp() *App {
	return &App{}
}

// Run creates objects (via constructors!)
func (a *App) Run(cfg *config.Config) {
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.ServerAddress)) // Помимо сообщения выведем параметр с адресом
	log.Debug("logger debug mode enabled")

	// Repository🧹🏦
	repo := storage.GetRepo(log, cfg)
	// //
	// repo, err := pg.NewPostgresRepo(log, cfg)
	// if err != nil {
	// 	log.Error("failed to connect storage")
	// 	os.Exit(1)
	// }

	// Use-Case🧹🏦
	// В данный момент именно service я не создаю. Сложно..
	// Видимо им можно считать вызов server.NewApp в main
	randomKey := random.RandomGenerator{}
	// Создаю сущность этого сервиса
	service := services.New(randomKey, repo, cfg)

	// HTTP Server🧹🏦
	// mux := chi.NewRouter()
	// // middlewares & handlers
	// v1.Router(mux, cfg, repo, randomKey, log)
	// a.httpServer = httpserver.New(cfg.HTTPServer.Address, mux, log)

	restServer, err := server.New(cfg, service)
	if err != nil {
		log.Error("failed to create http server", sl.Err(err))
		os.Exit(1)
	}

	// Waiting signal🧹🏦
	// done := make(chan os.Signal, 1)
	// signal.Notify(done, os.Interrupt, syscall.SIGINT, syscall.SIGTERM)

	// <-done
	// log.Info("stopping server")

	// // Смысл таймаута был, но сейчас потерян..
	// ctx := context.Background() //context.WithTimeout(context.Background(), 10*time.Second)
	// //defer cancel()

	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	wg.Add(2) //nolint:gomnd

	// Старт в двух горутинах
	go runServer(ctx, wg, restServer, "REST HTTP server", log)
	wg.Wait()

	// Shutdown🧹🏦
	// if err := a.httpServer.Shutdown(ctx); err != nil {
	// 	log.Error("failed to stop server", sl.Err(err))
	// 	return
	// }
	log.Info("trying to shutdown storage gracefully")

	errClose := repo.Close(context.Background()) //nolint:contextcheck
	if errClose != nil {
		//log.Fatal().Err(errClose)
		log.Error("the storage was NOT closed correctly", sl.Err(errClose))
		os.Exit(1)
	} else {
		log.Info("the storage was closed")
	}

	// TODO: close storage

	log.Info("server stopped")
}

func runServer(ctx context.Context, wg *sync.WaitGroup, server server.Server, serverName string, log *slog.Logger) {
	log.Info("%s started", serverName)

	go func() {
		<-ctx.Done()
		log.Info("trying to shutdown %s gracefully", serverName)

		if errShutdown := server.Shutdown(); errShutdown != nil {
			log.Info("%s server Shutdown: %v", serverName, errShutdown)
		} else {
			log.Info("%s shutted down gracefully", serverName)
		}
		wg.Done()
	}()

	if errRun := server.Run(); errRun != http.ErrServerClosed && errRun != nil {
		log.Info("%s could not have started: %v", serverName, errRun)
	}
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
