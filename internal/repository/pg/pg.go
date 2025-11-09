package pg

import (
	"app/internal/config"
	"app/internal/entity"
	"app/internal/usecase/logger/sl"
	"context"
	"database/sql"
	"fmt"
	"log/slog"
	"os"
	"time"

	// Импорт для "побочного эффекта" (side effect)
	// Значит, что нужны не все (основные) "эффекты" пакета,
	// а только дополнительные, нужные другим пакетам
	// (тут пакету "database/sql")
	_ "github.com/lib/pq"
)

type PostgresRepo struct {
	DB *sql.DB
}

// Инициализация подключения к PostgreSQL
func NewPostgresRepo(log *slog.Logger, cfg *config.Config) (*PostgresRepo, error) {
	// Getting DSN from environment variables
	//dsn := os.Getenv("DATABASE_DSN")

	// // dsn проверяем перед вызовом InitDB
	// if dsn == "" {
	// 	log.Error("DATABASE_DSN not specified in env")
	// 	//os.Exit(1)
	// }

	// 1. Подключение к базе
	db, err := sql.Open("postgres", cfg.DSN)
	if err != nil {
		log.Error("DB connection error", sl.Err(err))
		return nil, fmt.Errorf("connection error: %v", err)
	}

	// // Не забыть про defer!!
	// defer db.Close()

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяю подключение с таймаутом ответа
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// освобождаем ресурс
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error("error to ping", sl.Err(err))
		return nil, fmt.Errorf("error to ping db: %v", err)
	}

	repo := &PostgresRepo{DB: db}

	err = checkTab(log, repo)
	if err != nil {
		log.Error("failed to init storage")
		os.Exit(1)
	}

	return repo, nil
}

// Создаем таблицу, если ее еще нет
func checkTab(log *slog.Logger, repo *PostgresRepo) error {

	stmt, err := repo.DB.Prepare(`
	CREATE TABLE IF NOT EXISTS aliases(
        alias VARCHAR NOT NULL UNIQUE,
        url TEXT NOT NULL);
	`)

	if err != nil {
		log.Error(err.Error())
	}

	// Отправляем комманду (CREATE TABLE в данном случае)
	// Exec выполняет подготовленный оператор (stmt) с заданными аргументами
	// и возвращает [Result], суммирующий эффект оператора.
	// В данной ситуации этот "эффект" не используется
	_, err = stmt.Exec()
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}

func CreateRecord(log *slog.Logger, shortUrl entity.ShortURL, repo *PostgresRepo) error {
	const op = "repository.pg.CreateRecord" // Имя текущей функции для логов и ошибок
	url := shortUrl.OriginalURL
	alias := shortUrl.ID
	stmt, err := repo.DB.Prepare("INSERT INTO aliases(alias, url) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias, url)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}
	//

	return nil
}
