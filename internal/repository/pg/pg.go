package pg

import (
	"app/internal/config"
	"app/internal/domain/link"
	err_repo "app/internal/errors/repository"
	"app/pkg/logger/sl"
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/lib/pq"
)

type PostgresRepo struct {
	DB *sql.DB
}

func NewPostgresRepo(log *slog.Logger, cfg *config.Config) (*PostgresRepo, error) {
	// // DSN from environment variables
	// dsn := os.Getenv("DATABASE_DSN")

	// 1. Подключение к базе
	db, err := sql.Open("postgres", cfg.DatabaseDSN)
	if err != nil {
		log.Error("DB connection error", sl.Err(err))
		return nil, fmt.Errorf("connection error: %v", err)
	}
	// defer db.Close()

	// Настройки пула соединений
	db.SetMaxOpenConns(25)
	db.SetMaxIdleConns(25)
	db.SetConnMaxLifetime(5 * time.Minute)

	// Проверяю подключение с таймаутом ответа
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	// Всегда отменяем контекст, чтобы освободить его ресурсы
	defer cancel()

	if err := db.PingContext(ctx); err != nil {
		log.Error("error to ping db", sl.Err(err))
		return nil, err //fmt.Errorf("error to ping db: %v", err)
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
        id VARCHAR NOT NULL UNIQUE,
        url TEXT NOT NULL UNIQUE,
		correlation_id VARCHAR);
	`)

	if err != nil {
		log.Error(err.Error())
	}

	// Отправляем комманду (CREATE TABLE в данном случае)
	// Exec выполняет подготовленный оператор (stmt) с заданными аргументами
	// и возвращает [Result], суммирующий эффект оператора.
	// В данной ситуации это "побочный эффект" (не используется)
	_, err = stmt.Exec()
	if err != nil {
		log.Error(err.Error())
	}

	return nil
}

// Save проверяет уникальность URL-адреса и сохраняет его
func (repo *PostgresRepo) Save(ctx context.Context, shortURL link.ExpandedURL) error {
	const op = "repository.pg.Save" // Имя текущей функции для логов и ошибок
	url := shortURL.OriginalURL
	alias := shortURL.ID
	stmt, err := repo.DB.Prepare("INSERT INTO aliases(id, url) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias, url)

	// // Для устранения глубокой (4️⃣ уровня) вложенности if ... else
	// // использую паттерн ❗early return (ранний возврат)
	// // 1️⃣ Проверяем основную ошибку выполнения INSERT
	if err == nil {
		// "Ранний" возврат. Успешная вставка.
		return nil
	}

	var pqErr *pq.Error

	// // 2️⃣ Ошибка произошла. Проверяем, является ли она ошибкой PostgreSQL.
	// if errors.As(err, &pqErr) {
	if !errors.As(err, &pqErr) {
		// "Ранний" возврат. Это не ошибка pq (например, ошибка сети или таймаут).
		return fmt.Errorf("unknown database error: %w", err)
	}

	// iter13. Проверка на уникальность
	// 3️⃣ Это ошибка pq. Проверяем код ошибки.
	// Код ошибки "duplicate key value" or "UniqueViolation (?)"- "23505" (pgerrcode.UniqueViolation)
	if pqErr.Code != "23505" {
		// "Ранний" возврат. Это другая ошибка pq (например, нарушение NOT NULL).
		return fmt.Errorf("SQL code error %s: %w", pqErr.Code, err)
	}

	// Конфликт уникальности. Делаем дополнительный SELECT
	var existingID string
	selectStatement := `SELECT id FROM aliases WHERE url = $1`
	errSelect := repo.DB.QueryRow(selectStatement, url).Scan(&existingID)

	// 4️⃣ Последний "if", "освобожденный" от остальных,
	// стал соответствовать early return, даже без изменений
	if errSelect != nil {
		// Если SELECT тоже не сработал, возвращаем ошибку SELECT
		return fmt.Errorf("ошибка при получении существующего ID: %w", errSelect)
	}

	// Возвращаем пользовательскую ошибку с найденным ID
	return &err_repo.NotUniqueURLError{
		Err: nil,
		ShortURL: link.ExpandedURL{
			OriginalURL: url,
			ID:          existingID,
		},
	}
}

// SaveBatch сохраняет несколько URL-адресов.
// Проверяет уникальность URL-адресов и сохраняет их.
func (repo *PostgresRepo) SaveBatch(ctx context.Context, batch []link.ExpandedURL) error {
	// 1. Начинаем транзакцию с контекстом
	// Это гарантирует, что все операции COPY выполняются в рамках одного соединения.
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Откат транзакции, если произойдет ошибка до Commit()

	// 2. Подготавливаем операцию COPY IN
	// Указываем имя таблицы и список столбцов в целевой таблице БД.
	// Таблица называется 'aliases' с соответствующими столбцами.
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("aliases", "id", "url", "correlation_id"))
	if err != nil {
		return fmt.Errorf("failed to prepare COPY statement: %w", err)
	}
	defer stmt.Close()

	// 3. Пакетная вставка данных
	for _, url := range batch {
		// Используем ExecContext для передачи данных построчно
		_, err = stmt.ExecContext(ctx, url.ID, url.OriginalURL, url.CorrelationID)
		if err != nil {
			// В случае ошибки ExecContext автоматически вызовет Rollback() для стейтмента.
			return fmt.Errorf("failed to exec data row: %w", err)
		}
	}

	// 4. Завершаем поток данных (закрываем COPY)
	_, err = stmt.ExecContext(ctx)
	if err != nil {
		return fmt.Errorf("failed to close COPY data stream: %w", err)
	}

	// 5. Коммитим транзакцию
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

func (repo *PostgresRepo) Check(ctx context.Context) error {
	// и вся проверка "здоровья"!
	return repo.DB.PingContext(ctx)
}

// FindByID находит URL по идентификатору.
func (repo *PostgresRepo) FindByID(ctx context.Context, id string) (link.ExpandedURL, error) {
	var ent link.ExpandedURL
	err := repo.DB.QueryRowContext(
		ctx,
		"select url, id from aliases where id=$1",
		id,
	).Scan(&ent.OriginalURL, &ent.ID)
	return ent, err
}

// Stub function
func (repo *PostgresRepo) Close(_ context.Context) error {
	return nil
}
