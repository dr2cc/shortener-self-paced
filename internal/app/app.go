// Package app configures and runs application.
package app

import (
	"app/internal/config"
	"app/internal/handler"
	storage "app/internal/repository"
	"app/internal/server"
	"app/internal/service"
	"app/pkg/logger/sl"
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
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

	// ❗ В достижении цели “разделения ответственности между всеми слоями приложения” нам помогает “правило
	// зависимости” (это о круговой диаграмме дяди Боба).
	// Зависимости направлены только внутрь (внутренний круг ничего не должен знать про внешний
	// и сущности внутреннего круга не могут обратиться к сущностям внешнего).
	// ❗ И вот чтобы реализовать “Правило зависимости” мы используем технику dependency injection !

	// Создаем сущности слоев (это three-layered architecture)
	// в порядке обратном обращению к ним:
	//
	// 3️⃣ Repository🧹🏦 (DAL)
	// Создаем объект хранилища, в соответствии с настройками
	repository := storage.NewRepository(log, cfg)
	// ↑
	// | внедряем в бизнес-логику
	// Use-Case🧹🏦
	// Считаю, что здесь правильно присвоено значение
	// структуры RandomStringGenerator (по сути поведение- метод GenerateIDfromString)
	// а не интерфейса IDGenerator () (интерфейс служит границей между слоями)
	// randomKey := service.RandomStringGenerator{}
	// ↑
	// 2️⃣ Use case (BL - Business Logic Layer, service)
	// | Здесь внедряем зависимость с repository
	// ❌ (07.01.26) Убрать такие знаки в service!
	services := service.NewService(repository, cfg)
	// ↑
	// 1️⃣ Handler (PL - Presentation Layer, controller)
	// | Здесь внедряем зависимость с services
	handlers := handler.NewHandler(services)
	// ↑ HTTP request

	// Работает в обратном раправлении!
	// HTTP запрос -> ручка -> обращение к службе -> служба к базе данных.

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

	// Отдельная горутина: сервер запускается в своей собственной горутине.
	// Это необходимо, так как ListenAndServe() является блокирующим вызовом.
	go func() {
		if err := srv.Run(cfg.ServerAddress, handlers.InitRoutes(log)); err != nil {
			// В Go метод http.Server.ListenAndServe() ("спрятан" внутри srv.Run()) спроектирован так, что он всегда возвращает ошибку, если работа прекращена.
			// Если это "авария" — вернется реальная ошибка.
			// Если вы сами вызвали srv.Shutdown() — вернется специальная переменная http.ErrServerClosed.
			// Если её не «отфильтровать», то приложение при выключении всегда будет "бросать" Fatal или os.Exit(1),
			// что ломает тесты и запутывает системы мониторинга (они считают, что сервис упал, а не выключился).

			// Проверяем, что ошибка НЕ является нашим сигналом о закрытии сервера
			if !errors.Is(err, http.ErrServerClosed) {
				// Логируем как Critical или Error, так как сервер не смог запуститься или упал
				log.Error("server listener crashed", sl.Err(err))
				// Важно: закрываем основной процесс, так как без сервера приложение бесполезно
				os.Exit(1)
			}
			// Если ошибка == ErrServerClosed, мы просто выходим из горутины.
			// Это нормальное поведение при завершении.
		}
	}()

	log.Info("ShortenerApp is started")

	// Graceful shutdown
	// quit: Это наш "стоп-кран".
	// Это буферизованный канал, который будет ожидать системные сигналы.
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	log.Info("ShortenerApp is shutting down")

	// Корректное завершение (?)
	// Используем корневой контекст Background
	if err := srv.Shutdown(context.Background()); err != nil {
		log.Error("failed to create http server", sl.Err(err))
		os.Exit(1)
	}

	// // TODO: Close storage
	// if err := db.Close(); err != nil {
	// 	// Я не использую логгер logrus
	// 	// logrus.Errorf("error occured on db connection close: %s", err.Error())
	// 	log.Error("error occured on db connection close", sl.Err(err))
	// 	os.Exit(1)
	// }
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
