// Package app configures and runs application.
package app

import (
	"app/internal/config"
	"app/internal/generator"
	"app/internal/handler"
	storage "app/internal/repository"
	"app/internal/server"
	"app/internal/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

const (
	envLocal = "local"
	envDev   = "dev"
	envProd  = "prod"
)

func Run(cfg *config.Config) error {
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.ServerAddress))

	// 3️⃣ Repository🧹🏦 (DAL)
	repository := storage.NewRepository(log, cfg)
	// 2️⃣ ❌Use case (BL - Business Logic Layer, service)
	generator := generator.NewStringGenerator()
	services := service.NewService(repository, generator, cfg)
	// 1️⃣ Handler (PL - Presentation Layer, controller)
	handlers := handler.NewHandler(services)
	// ↑ HTTP request

	// HTTP Server🧹🏦
	srv := new(server.Server)

	// Первоначальный запуск и остановка сервера из todo-app1
	// Но он не проходил тесты yp❗ Причина:
	// Внутри if err := srv.Run(); err != nil{} («использование if с коротким объявлением (short statement)
	// для немедленной обработки ошибки запуска сервера») происходят две вещи:
	// 1. Печатается лог.
	// 2. Вызывается os.Exit(1).
	// В результате, когда тест останавливает сервер, приложение вместо «чистого» выхода (статус 0)
	// принудительно завершается с ошибкой (статус 1).
	// Тестовый сьют видит этот статус и считает, что сервер «упал».
	//
	// Добавил проверку //if !errors.Is(err, http.ErrServerClosed) {})// и убрал //os.Exit(1)// (он мешал обработке ошибки запуска в main).

	// Канал для ошибок сервера
	serverErrors := make(chan error, 1)
	// Сервер запускается в отдельной горутине (ListenAndServe() является блокирующим вызовом).
	go func() {
		log.Info("ShortenerApp is starting", slog.String("addr", cfg.ServerAddress))

		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes(log)); err != nil {
			// http.Server.ListenAndServe() всегда возвращает ошибку, если работа прекращена.
			// Если это "авария" — вернется реальная ошибка.
			// Если вызван srv.Shutdown() — вернется специальная переменная http.ErrServerClosed.
			// Если её не обработать, то приложение при выключении всегда будет "бросать" Fatal или os.Exit(1),
			// что ломает тесты и запутывает системы мониторинга (они считают, что сервис упал, а не выключился).

			// Проверяем, что ошибка НЕ является нашим сигналом о закрытии сервера
			if !errors.Is(err, http.ErrServerClosed) {
				// Вместо os.Exit отправляем ошибку в канал
				serverErrors <- fmt.Errorf("server listener crashed: %w", err)
				// os.Exit(1)
			}
			// Если ошибка == ErrServerClosed, выходим из горутины.
		}
	}()

	// Graceful shutdown
	// quit: Это "стоп-кран". Это буферизованный канал, который будет ожидать системные сигналы.
	quit := make(chan os.Signal, 1)
	// Указываем сигналы, которые хотим слушать
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)

	// select для ожидания либо сигнала, либо ошибки сервера
	select {
	case err := <-serverErrors:
		return err // Возвращаем ошибку, если сервер упал сам

	case sig := <-quit:
		log.Info("ShortenerApp is shutting down", slog.String("signal", sig.String()))
		//log.Info("ShortenerApp is shutting down")

		// ВАЖНО: после получения первого сигнала, остальные не слушаем. Это вернет стандартное поведение системы.
		signal.Stop(quit)

		// Контекст с таймаутом для Shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return fmt.Errorf("failed to shutdown http server: %w", err)
		}
	}

	// // TODO: Close storage
	// if err := db.Close(); err != nil {
	// 	// Я не использую логгер logrus
	// 	// logrus.Errorf("error occured on db connection close: %s", err.Error())
	// 	log.Error("error occured on db connection close", sl.Err(err))
	// 	os.Exit(1)
	// }

	return nil
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
