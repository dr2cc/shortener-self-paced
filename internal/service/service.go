// The service package contains the core business logic of the application.
package service

import (
	"app/internal/config"
	"app/internal/entity"
	storage "app/internal/repository"
	"context"
	"errors"
	"fmt"
	"time"
)

// // TODO❗ ShortenerInterface со всем поведением службы Shortener,
// // будет необходим для сервера gRPC
// type ShortenerInterface interface {
// ShortenBatch(ctx context.Context, batch []entity.ExpandedURL, userID string) ([]entity.ExpandedURL, error)
// Shorten(ctx context.Context, url string, userID string) (entity.ExpandedURL, error)
// FindURL(ctx context.Context, id string) (entity.ExpandedURL, error)
// HealthCheck(ctx context.Context) error
// FormatShortURL(urlID string) string
// }

// Здесь определены предметные области (доменные зоны).
// ❗Предметная область это круг задач (сферы реального мира) решаемых приложением.
// Предметные области этого проекта (по мере добавления):
// 🔸 сокращение URL.
//
// Предотвращение «утечек абстракции» https://habr.com/ru/articles/881918/
type Service struct {
	// Сервис сокращения URL (создает ID (shortURL) из url), со своим функционалом
	Random IDGenerator
	// ❌ по todo-app, "общаться" с хранилищем правильнее из самого сервиса (пока только Random)
	repository storage.Repository
	config     *config.Config
	//generator  generator.URLGenerator
}

// Вызывается из app
func NewService(rand IDGenerator, repo storage.Repository, conf *config.Config) *Service {
	return &Service{
		Random: rand,
		// // ❌ по todo-app вот так создают сервисы приложения
		// // Создаю новый проект. Назову shortener-todo-app. В нем iter1 и затем iter13(!?)
		// // В iter1 привожу к todo-app (м.б. и gin использовать?!)
		// Random:     NewRandomService(repo.Random),
		repository: repo,
		config:     conf,
		//generator:  generator,
	}
}

// ShortenBatch сокращает массив значений []entity.ExpandedURL
// Все записи пакета должны содержать OriginalURL.
func (sh *Service) ShortenBatch(ctx context.Context, batch []entity.ExpandedURL) ([]entity.ExpandedURL, error) {
	for i, URL := range batch {
		urlID, err := sh.Random.GenerateIDfromString(URL.OriginalURL)
		if err != nil {
			return nil, err
		}
		batch[i].ID = urlID
		//batch[i].CreatedByID = userID
	}

	if err := sh.repository.SaveBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// Shorten сокращает полный URL и возвращает заполненную структуру ShortURL
func (sh *Service) Shorten(ctx context.Context, url string) (entity.ExpandedURL, error) {
	urlID, err := sh.Random.GenerateIDfromString(url)
	if err != nil {
		return entity.ExpandedURL{}, err
	}

	shortURL := entity.ExpandedURL{
		OriginalURL: url,
		ID:          urlID,
		// CreatedByID: userID,
	}

	// Пробуем записать в хранилище, с проверкой уникальности (iter13)
	err = sh.repository.Save(ctx, shortURL)

	// "Реакция" на уникальность / не уникальность
	var notUniqueErr *storage.NotUniqueURLError
	// 🔦 func errors.As(err error, target any) bool
	// As находит первую ошибку в дереве err, соответствующую target, и, если она найдена,
	// устанавливает target равным этому значению ошибки и возвращает true.
	// В противном случае возвращает false.
	// Дерево состоит из самого err, за которым следуют ошибки,
	// полученные путём многократного вызова его метода Unwrap() error или Unwrap() []error.
	// Когда err оборачивает несколько ошибок, As проверяет err, а затем выполняет обход в глубину его дочерних элементов.
	if errors.As(err, &notUniqueErr) {
		// Если запись не уникальна, возвращаем:
		return shortURL, NewShorteningError(shortURL, err)
	}

	if err != nil {
		return entity.ExpandedURL{}, err
	}

	return shortURL, nil
}

// Функция FindURL находит в хранилище полный URL-адрес по указанному идентификатору.
// Возвращает заполненную структуру entity.ExpandedURL
func (sh *Service) FindURL(ctx context.Context, id string) (entity.ExpandedURL, error) {
	origURL, err := sh.repository.FindByID(ctx, id)
	if err != nil {
		return entity.ExpandedURL{}, err //Shortener
	}
	return origURL, nil
}

// HealthCheck проверяет корректность работы выбранного хранилища
func (sh *Service) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return sh.repository.Check(ctx)
}

// FormatShortURL форматирует полученный идентификатор URL
// в результирующую строку, возвращаемую запросами POST
func (sh *Service) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", sh.config.BaseURL, urlID)
}

// shorteningError — это обертка (wrapper) любой ошибки,
// возникшей в процессе работы службы
type shorteningError struct {
	Err      error
	ShortURL entity.ExpandedURL
}

func (err *shorteningError) Error() string {
	return fmt.Sprintf("error while shortening: %v", err.Err)
}

func (err *shorteningError) Unwrap() error {
	return err.Err
}

// NewShorteningError добавляет (wraps) к ошибке поле err с дополнительной информацией об URL
func NewShorteningError(shortURL entity.ExpandedURL, err error) error {
	return &shorteningError{
		Err:      err,
		ShortURL: shortURL,
	}
}
