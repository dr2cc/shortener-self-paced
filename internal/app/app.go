package app

import (
	"app/internal/config"
	storage "app/internal/repository"
	"app/internal/server"
	"app/internal/usecase/logger/sl"
	"app/internal/usecase/random"
	services "app/internal/usecase/shortener"
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

// Run creates objects (via constructors!)
func Run(cfg *config.Config) {
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
	randomKey := random.RandomGenerator{}
	// Создаю сущность этого сервиса
	service := services.New(randomKey, repo, cfg)

	// HTTP Server🧹🏦
	httpServer, err := server.New(cfg, service)
	if err != nil {
		log.Error("failed to create http server", sl.Err(err))
		os.Exit(1)
	}

	// Waiting signal🧹🏦
	ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	wg := &sync.WaitGroup{}
	// TODO: готовлюсь к двум горутинам при добавлении сервера grpc
	// тогда будет wg.Add(2)
	wg.Add(1) //nolint:gomnd

	go runServer(ctx, wg, httpServer, "HTTP server", log)
	// // когда добавлю grpc, то добавлю такую строку:
	// go runServer(ctx, wg, grpcServer, "GRPC server")

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
