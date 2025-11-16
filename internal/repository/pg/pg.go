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

	"github.com/lib/pq"
	_ "github.com/lib/pq"
	// _ {import} это импорт для "побочного эффекта" (side effect)
	// Значит, что нужны не все (основные) "эффекты" пакета,
	// а только дополнительные, нужные другим пакетам
	// (тут пакету "database/sql" нужен импорт драйвера)
)

type PostgresRepo struct {
	DB *sql.DB
}

// Инициализация подключения к PostgreSQL
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
        url TEXT NOT NULL,
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

func (repo *PostgresRepo) Save(ctx context.Context, shortURL entity.ShortURL) error {
	const op = "repository.pg.Save" // Имя текущей функции для логов и ошибок
	url := shortURL.OriginalURL
	alias := shortURL.ID
	stmt, err := repo.DB.Prepare("INSERT INTO aliases(alias, url) VALUES($1, $2)")
	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	_, err = stmt.Exec(alias, url)

	if err != nil {
		return fmt.Errorf("%s: %w", op, err)
	}

	return nil
}

// func (repo *PostgresRepo) SaveBatch(ctx context.Context, batch []entity.ShortURL) error {
// 	//_, err := repo.conn.CopyFrom(
// 	_, err := repo.DB.PrepareContext(ctx, pq.CopyIn(
// 		"aliases",
// 		[]string{"original_url", "id", "created_by", "correlation_id", "deleted_at"},
// 		pgx.CopyFromSlice(len(batch), func(i int) ([]interface{}, error) {
// 			return []interface{}{batch[i].OriginalURL, batch[i].ID, batch[i].CreatedByID, batch[i].CorrelationID, batch[i].DeletedAt}, nil
// 		}),
// 	)
// 	return err
// }

// Образец от G
// func BulkInsertShortURLs(ctx context.Context, db *sql.DB, urls []ShortURL) error {
func (repo *PostgresRepo) SaveBatch(ctx context.Context, batch []entity.ShortURL) error {
	// 1. Начинаем транзакцию с контекстом
	// Это гарантирует, что все операции COPY выполняются в рамках одного соединения.
	tx, err := repo.DB.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback() // Откат транзакции, если произойдет ошибка до Commit()

	// 2. Подготавливаем операцию COPY IN
	// Указываем имя таблицы и список столбцов в целевой таблице БД.
	// Таблица называется 'short_urls' с соответствующими столбцами.
	stmt, err := tx.PrepareContext(ctx, pq.CopyIn("aliases", "alias", "url", "correlation_id"))
	//stmt, err := repo.DB.PrepareContext(ctx, pq.CopyIn("aliases", "alias", "url", "correlation_id"))
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
	return repo.DB.PingContext(ctx)
}

func (repo *PostgresRepo) GetByID(ctx context.Context, id string) (entity.ShortURL, error) {
	var ent entity.ShortURL
	err := repo.DB.QueryRowContext(
		ctx,
		"select url, alias from aliases where alias=$1",
		id,
	).Scan(&ent.OriginalURL, &ent.ID)
	return ent, err
}

// Stub function
func (repo *PostgresRepo) Close(_ context.Context) error {
	return nil
}
