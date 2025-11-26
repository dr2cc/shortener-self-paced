// Package app configures and runs application.
package app

import (
	"app/internal/config"
	"app/internal/handlers"
	"app/internal/server"
	"app/internal/storage"
	"app/internal/storage/cache"
	jsonstore "app/internal/storage/jsonrstore"
	"app/internal/storage/pg"
	"app/internal/usecase/random"
	services "app/internal/usecase/shortener"
	"app/pkg/logger/sl"
	"context"
	"fmt"
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

// Run создает объекты (через конструкторы!)
func Run(cfg *config.Config) {
	// Создаем объект логгера
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.ServerAddress))
	log.Debug("logger debug mode enabled")

	// Repository🧹🏦
	// Создаем объект хранилища, в соответствии с настройками
	repo := choosingStorage(log, cfg)

	// Use-Case🧹🏦
	// Считаю, что здесь правильно присвоено значение
	// структуры RandomStringGenerator (по сути поведение- метод GenerateIDfromString)
	// а не интерфейса IDGenerator () (интерфейс служит границей между слоями)
	randomKey := random.RandomStringGenerator{}
	// Создаем "сущность" этого сервиса
	// TODO👀 - спросить наставника, в чем смысл такой сущности (еще глянуть в обеих чистых архитектурах)
	// По моему мнению- чтобы в любом месте проекта были доступны основные методы именно из этй сущности,
	// а не напрямую (разделение слоев?)
	service := services.New(randomKey, repo, cfg)

	// HTTP Server🧹🏦
	router := handlers.NewRouter(service, cfg, log)
	restAPIserver, err := server.New(cfg, router)
	if err != nil {
		log.Error("failed to create http server", sl.Err(err))
		os.Exit(1)
	}

	// Waiting signal🧹🏦
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	// TODO: готовлюсь к двум горутинам при добавлении сервера grpc
	// тогда будет wg.Add(2)
	wg.Add(1)

	go runServer(ctx, wg, restAPIserver, "REST API server", log)
	// // когда добавлю gRPC, то добавлю такую строку:
	// go runServer(ctx, wg, grpcAPIserver, "gRPC API server", log)

	wg.Wait()

	// // TODO: close storage
	// log.Info("trying to shutdown storage")
	// errClose := repo.Close(context.Background()) //nolint:contextcheck
	// if errClose != nil {
	// 	//log.Fatal().Err(errClose)
	// 	log.Error("the storage was NOT closed correctly", sl.Err(errClose))
	// 	os.Exit(1)
	// } else {
	// 	log.Info("the storage was closed")
	// }
}

func choosingStorage(log *slog.Logger, cfg *config.Config) storage.Repository {
	if cfg.DatabaseDSN != "" {
		repo, err := pg.NewPostgresRepo(log, cfg)
		if err != nil {
			log.Error("failed to connect pg storage")
			os.Exit(1)
		}
		return repo
	}
	if cfg.FilePath != "" {
		repo, err := jsonstore.NewFileRepository(cfg.FilePath)
		if err != nil {
			log.Error("file (jsonstore) storage error")
			os.Exit(1)
		}
		return repo
	}

	return cache.NewInMemoryRepository()
}

func runServer(ctx context.Context, wg *sync.WaitGroup, server server.Server, servName string, log *slog.Logger) {
	log.Info(fmt.Sprintf("%s started", servName))

	go func() {
		<-ctx.Done()
		// Shutdown🧹🏦+
		log.Info(fmt.Sprintf("trying to stop %s", servName))
		if errShutdown := server.Shutdown(); errShutdown != nil {
			log.Info("%s server shutdown: %v", servName, errShutdown)
		} else {
			log.Info(fmt.Sprintf("%s stopped", servName))
		}
		wg.Done()
	}()

	if errRun := server.Run(); errRun != http.ErrServerClosed && errRun != nil {
		log.Info("%s could not have start: %v", servName, errRun)
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
