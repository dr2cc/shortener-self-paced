// The services package contains the core business logic of the application.
package services

import (
	"app/internal/config"
	"app/internal/entity"
	"app/internal/storage"
	"app/internal/usecase/generator"
	"app/internal/usecase/random"
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

// Shortener — служба, предоставляющая хранилище, бизнес-логику, конфигурацию.
// Обращается она к хранилищу и бизнес-логике исключительно через интерфейсы.
// Это соответствует подходу, в котором "внешний мир" для usecase
//
//	— это всего лишь набор интерфейсов,
//
// которые оперируют сущностями из слоя domain и не имеют никаких деталей о том,
// кто и как реализует эти интерфейсы.
// Но shortener выполняет не единственную задачу и в этом не соответствует usecase.
// ❗Фундаментальное отличие между service и usecase заключается в том, что
// service, как правило, реализует набор методов, тогда как
// usecase сосредоточен на выполнении одной конкретной задачи,
// обеспечивая высокую изоляцию и строгое соблюдение принципа единственной ответственности.
// В итоге service, который в процессе рефакторинга стремится к более высокой степени изоляции
//
//	и всё больше ориентируется на соблюдение принципа единственной ответственности, рано или поздно вырождается в usecase.
//
// https://habr.com/ru/articles/881918/
type Shortener struct {
	repository storage.Repository
	generator  generator.URLGenerator
	Random     random.Generator
	config     *config.Config
}

// В этом конструкторе создаем службу сокращения URL - Shortener
func New(repo storage.Repository, generator generator.URLGenerator, random random.Generator, conf *config.Config) *Shortener {
	return &Shortener{
		repository: repo,
		generator:  generator,
		Random:     random,
		config:     conf,
	}
}

// ShortenBatch сокращает массив значений типа []entity.ExpandedURL
// Все записи пакета должны содержать OriginalURL.
func (sh *Shortener) ShortenBatch(ctx context.Context, batch []entity.ExpandedURL, userID string) ([]entity.ExpandedURL, error) {
	for i, URL := range batch {
		urlID, err := sh.generator.GenerateIDFromString(URL.OriginalURL)
		if err != nil {
			return nil, err
		}
		batch[i].ID = urlID
		batch[i].CreatedByID = userID
	}

	if err := sh.repository.SaveBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// Shorten сокращает полный URL и возвращает заполненную структуру ShortURL
func (sh *Shortener) Shorten(ctx context.Context, url string, userID string) (entity.ExpandedURL, error) {
	urlID, err := sh.generator.GenerateIDFromString(url)
	if err != nil {
		return entity.ExpandedURL{}, err
	}

	shortURL := entity.ExpandedURL{
		OriginalURL: url,
		ID:          urlID,
		CreatedByID: userID,
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
func (sh *Shortener) FindURL(ctx context.Context, id string) (entity.ExpandedURL, error) {
	origURL, err := sh.repository.FindByID(ctx, id)
	if err != nil {
		return entity.ExpandedURL{}, err //Shortener
	}
	return origURL, nil
}

// HealthCheck проверяет корректность работы выбранного хранилища
func (sh *Shortener) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return sh.repository.Check(ctx)
}

// FormatShortURL форматирует полученный идентификатор URL
// в результирующую строку, возвращаемую запросами POST
func (sh *Shortener) FormatShortURL(urlID string) string {
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

// GenerateNewUserID generates new user id.
// It's just a wrapper for random.GenerateNewUserID().
func (sh *Shortener) GenerateNewUserID() string {
	//return sh.Random.GenerateNewUserID()
	return sh.Random.GenerateNewUserID()
}

// GetUrlsCreatedBy returns array of all urs that was shortened by given userID.
// It's just a wrapper for repository.GetUsersUrls.
// GetUrlsCreatedBy возвращает массив всех URL-адресов, сокращённых по заданному идентификатору пользователя.
// Это только оболочка для repository.GetUsersUrls (в каждом из видов хранилищ).
func (sh *Shortener) GetUrlsCreatedBy(ctx context.Context, userID string) ([]entity.ExpandedURL, error) {
	return sh.repository.GetUsersUrls(ctx, userID)
}
