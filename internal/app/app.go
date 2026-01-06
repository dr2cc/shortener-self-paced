// Package app configures and runs application.
package app

import (
	"app/internal/config"
	"app/internal/handler"
	storage "app/internal/repository"
	"app/internal/repository/cache"
	jsonstore "app/internal/repository/jsonrstore"
	"app/internal/repository/pg"
	"app/internal/server"
	"app/internal/service"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/sirupsen/logrus"
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

	// Создаем сущности слоев в обратном порядке!
	//
	// 3️⃣ Repository🧹🏦 (DAL)
	// Создаем объект хранилища, в соответствии с настройками
	repository := choosingStorage(log, cfg)
	// ↑
	// | внедряем в бизнес-логику
	// Use-Case🧹🏦
	// Считаю, что здесь правильно присвоено значение
	// структуры RandomStringGenerator (по сути поведение- метод GenerateIDfromString)
	// а не интерфейса IDGenerator () (интерфейс служит границей между слоями)
	randomKey := service.RandomStringGenerator{}
	// Создаем "сущность" этого сервиса
	// В чем смысл такой сущности (еще глянуть в обеих чистых архитектурах)?
	// По моему мнению- чтобы в любом месте проекта были доступны основные методы именно из этй сущности,
	// а не напрямую (разделение слоев?)
	// Нет! Это и есть:
	// 2️⃣ Use case (BL)!
	services := service.NewService(randomKey, repository, cfg)
	// ↑
	// |

	// ❗ Начало из todo-app1
	// 1️⃣ Handler (PL - Presentation Layer, controller)
	// | Здесь внедряем зависимость с services
	handlers := handler.NewHandler(services)
	// ↑ HTTP request

	// Работает в обратном раправлении!
	// HTTP запрос -> ручка -> обращение к службе -> служба к базе данных.

	// HTTP Server🧹🏦
	srv := new(server.Server)

	// Отдельная горутина: сервер запускается в своей собственной горутине.
	// Это необходимо, так как ListenAndServe() является блокирующим вызовом.

	//// ❗Вариант из todo-app1 не проходил тесты.
	//// Причина:
	// 	Функция logrus.Fatalf делает две вещи:
	// Печатает лог.
	// Вызывает os.Exit(1).
	// В результате, когда тест останавливает сервер, ваше приложение вместо «чистого» выхода (статус 0)
	// принудительно завершается с ошибкой (статус 1).
	// Тестовый сьют видит этот статус и считает, что сервер «упал».
	//
	// Добавил проверку:
	// if !errors.Is(err, http.ErrServerClosed) {}

	go func() {
		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes(log)); err != nil {
			// Проверяем, что ошибка НЕ является сигналом о закрытии сервера
			if !errors.Is(err, http.ErrServerClosed) {
				logrus.Fatalf("error occured while running http server: %s", err.Error())
			}
		}
	}()

	logrus.Print("ShortenerApp Started")

	// Graceful shutdown
	// quit: Это наш "стоп-кран".
	// Это буферизованный канал, который будет ожидать системные сигналы.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	logrus.Print("ShortenerApp Shutting Down")

	// Корректное завершение (?)
	// Используем корневой контекст Background
	if err := srv.Shutdown(context.Background()); err != nil {
		logrus.Errorf("error occured on server shutting down: %s", err.Error())
	}

	// // TODO: Close storage
	// if err := db.Close(); err != nil {
	// 	logrus.Errorf("error occured on db connection close: %s", err.Error())
	// }
	// ❗ Конец из todo-app1

	// // // ❌ Переделать вызов сервера!! Как в todo-app1
	// // 1️⃣ Handler (PL)
	// router := handler.InitRoutes(services, cfg, log)

	// // HTTP Server🧹🏦
	// restAPIserver, err := server.New(cfg, router)
	// if err != nil {
	// 	log.Error("failed to create http server", sl.Err(err))
	// 	os.Exit(1)
	// }

	// // Waiting signal🧹🏦
	// ctx, _ := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	// wg := &sync.WaitGroup{}
	// // TODO: готовлюсь к двум горутинам при добавлении сервера grpc
	// // тогда будет wg.Add(2)
	// wg.Add(1)

	// go runServer(ctx, wg, restAPIserver, "REST API server", log)
	// // // когда добавлю gRPC, то добавлю такую строку:
	// // go runServer(ctx, wg, grpcAPIserver, "gRPC API server", log)

	// wg.Wait()

	// // // TODO: close storage
	// // log.Info("trying to shutdown storage")
	// // errClose := repo.Close(context.Background()) //nolint:contextcheck
	// // if errClose != nil {
	// // 	//log.Fatal().Err(errClose)
	// // 	log.Error("the storage was NOT closed correctly", sl.Err(errClose))
	// // 	os.Exit(1)
	// // } else {
	// // 	log.Info("the storage was closed")
	// // }
}

// func runServer(ctx context.Context, wg *sync.WaitGroup, server server.Server, servName string, log *slog.Logger) {
// 	log.Info(fmt.Sprintf("%s started", servName))

// 	go func() {
// 		<-ctx.Done()
// 		// Shutdown🧹🏦+
// 		log.Info(fmt.Sprintf("trying to stop %s", servName))
// 		if errShutdown := server.Shutdown(); errShutdown != nil {
// 			log.Info("%s server shutdown: %v", servName, errShutdown)
// 		} else {
// 			log.Info(fmt.Sprintf("%s stopped", servName))
// 		}
// 		wg.Done()
// 	}()

// 	if errRun := server.Run(); errRun != http.ErrServerClosed && errRun != nil {
// 		log.Info("%s could not have start: %v", servName, errRun)
// 	}
// }

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
