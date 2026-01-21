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

// Run создает объекты (через конструкторы!)
func Run(cfg *config.Config) error {
	// Создаем объект логгера
	log := setupLogger(cfg.Env)
	log.Info("init server", slog.String("address", cfg.ServerAddress))

	// Создаем сущности слоев (это three-layered architecture)
	// в порядке обратном обращению к ним:
	//
	// 3️⃣ Repository🧹🏦 (DAL)
	// Создаем объект хранилища, в соответствии с настройками
	repository := storage.NewRepository(log, cfg)
	// ↑
	// 2️⃣ Use case (BL - Business Logic Layer, service)
	// | Здесь внедряем зависимость с repository
	// ❌ (07.01.26) Убрать такие знаки в service!
	generator := generator.NewStringGenerator()
	services := service.NewService(repository, generator, cfg)
	// ↑
	// 1️⃣ Handler (PL - Presentation Layer, controller)
	// | Здесь внедряем зависимость с services
	handlers := handler.NewHandler(services)
	// ↑ HTTP request

	// HTTP Server🧹🏦
	srv := new(server.Server)

	// Запуск и остановку сервера беру из todo-app1
	// Но он не проходил тесты yp❗
	// Причина:
	// Внутри if err := srv.Run(); err != nil{} («использование if с коротким объявлением (short statement)
	// для немедленной обработки ошибки запуска сервера») происходят две вещи:
	// 1. Печатается лог.
	// 2. Вызывается os.Exit(1).
	// В результате, когда тест останавливает сервер, приложение вместо «чистого» выхода (статус 0)
	// принудительно завершается с ошибкой (статус 1).
	// Тестовый сьют видит этот статус и считает, что сервер «упал».
	//
	// Добавил проверку:
	// if !errors.Is(err, http.ErrServerClosed) {}

	// ♊1. Добавляем канал для ошибок сервера
	serverErrors := make(chan error, 1)
	go func() {
		log.Info("ShortenerApp is starting", slog.String("addr", cfg.ServerAddress))
		// Отдельная горутина: сервер запускается в своей собственной горутине.
		// Это необходимо, так как ListenAndServe() является блокирующим вызовом.
		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes(log)); err != nil {
			// В Go метод http.Server.ListenAndServe() ("спрятан" внутри srv.Run()) спроектирован так, что он всегда возвращает ошибку, если работа прекращена.
			// Если это "авария" — вернется реальная ошибка.
			// Если вы сами вызвали srv.Shutdown() — вернется специальная переменная http.ErrServerClosed.
			// Если её не «отфильтровать», то приложение при выключении всегда будет "бросать" Fatal или os.Exit(1),
			// что ломает тесты и запутывает системы мониторинга (они считают, что сервис упал, а не выключился).

			// Проверяем, что ошибка НЕ является нашим сигналом о закрытии сервера
			if !errors.Is(err, http.ErrServerClosed) {
				// ♊Вместо os.Exit отправляем ошибку в канал
				serverErrors <- fmt.Errorf("server listener crashed: %w", err)

				// // Логируем как Critical или Error, так как сервер не смог запуститься или упал
				// log.Error("server listener crashed", sl.Err(err))
				// // Важно: закрываем основной процесс, так как без сервера приложение бесполезно
				// os.Exit(1)
			}
			// Если ошибка == ErrServerClosed, мы просто выходим из горутины.
			// Это нормальное поведение при завершении.
		}
	}()

	//log.Info("ShortenerApp is started")

	// Graceful shutdown
	// quit: Это наш "стоп-кран".
	// Это буферизованный канал, который будет ожидать системные сигналы.
	quit := make(chan os.Signal, 1)
	// Указываем сигналы, которые хотим слушать
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	//<-quit

	// ♊2. Используем select для ожидания либо сигнала, либо ошибки сервера
	select {
	case err := <-serverErrors:
		return err // Возвращаем ошибку, если сервер упал сам

	case sig := <-quit:
		log.Info("ShortenerApp is shutting down", slog.String("signal", sig.String()))
		//log.Info("ShortenerApp is shutting down")

		// // Корректное завершение (?)
		// // Используем корневой контекст Background
		// if err := srv.Shutdown(context.Background()); err != nil {
		// 	log.Error("failed to create http server", sl.Err(err))
		// 	os.Exit(1)
		// }

		// ВАЖНО: Перестаем слушать сигналы сразу после получения первого.
		// Это вернет стандартное поведение системы (второй Ctrl+C просто убьет процесс)
		signal.Stop(quit)

		// ♊3. Используем контекст с таймаутом для Shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			// // Если не удалось закрыть красиво, Close() закроет принудительно
			// srv.Close()
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
