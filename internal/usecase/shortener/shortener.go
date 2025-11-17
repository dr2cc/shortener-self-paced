// Package services contains main business logic of application.
package services

import (
	"app/internal/config"
	"app/internal/entity"
	"app/internal/storage"
	"app/internal/usecase/random"
	"context"
	"errors"
	"fmt"
	"time"
)

type ShortenerInterface interface {
	// r.Post("/", h.Shorten)
	ShortText(ctx context.Context, url string, userID string) (entity.ShortURL, error)
	// r.Get("/{id}", h.Expand)
	Redirect(ctx context.Context, id string) (entity.ShortURL, error)
	// FormatShortURL(urlID string) string
	// GetUrlsCreatedBy(ctx context.Context, userID string) ([]entity.ShortURL, error)
	// HealthCheck(ctx context.Context) error
	// ShortenBatch(ctx context.Context, batch []entity.ShortURL, userID string) ([]entity.ShortURL, error)
	// GenerateNewUserID() string
	// DeleteUrls(ctx context.Context, ids []string, userID string)
	// GetStats(ctx context.Context) (entity.Stats, error)
}

// Shortener is the main service of the application
// Shortener — основной сервис приложения
type Shortener struct {
	Random     random.RandomGenerator
	repository storage.Repository
	//generator  generator.URLGenerator
	config *config.Config
}

// New creates new service.
func New(
	random random.RandomGenerator,
	repository storage.Repository,
	config *config.Config,
	//generator generator.URLGenerator,
) *Shortener {
	return &Shortener{
		Random:     random,
		repository: repository,
		config:     config,
		//generator:  generator,
	}
}

// shorteningError is error wrapper of any error occurred in service.
type shorteningError struct {
	Err      error
	ShortURL entity.ShortURL
}

func (err *shorteningError) Error() string {
	return fmt.Sprintf("error while shortening: %v", err.Err)
}

func (err *shorteningError) Unwrap() error {
	return err.Err
}

// NewShorteningError wraps error err with additional info about url.
func NewShorteningError(shortURL entity.ShortURL, err error) error {
	return &shorteningError{
		Err:      err,
		ShortURL: shortURL,
	}
}

// Shorten shortens full url and returns filled struct ShortURL.
func (service *Shortener) Shorten(ctx context.Context, url string) (entity.ShortURL, error) {
	// urlID, err := service.generator.GenerateIDFromString(url)
	// if err != nil {
	// 	return entity.ShortURL{}, err
	// }

	shortURL := entity.ShortURL{
		OriginalURL: url,
		ID:          service.Random.NewRandomString(),
		//ID:          urlID,
		// CreatedByID: userID,
	}

	err := service.repository.Save(ctx, shortURL)
	var notUniqueErr *storage.NotUniqueURLError
	if errors.As(err, &notUniqueErr) {
		return shortURL, NewShorteningError(shortURL, err)
	}
	if err != nil {
		return entity.ShortURL{}, err
	}

	return shortURL, nil
}

// Expand expands full url from given id. Returns filled ShortURL struct.
func (service *Shortener) Expand(ctx context.Context, id string) (entity.ShortURL, error) {
	origURL, err := service.repository.GetByID(ctx, id)
	if err != nil {
		return entity.ShortURL{}, err
	}
	return origURL, nil
}

// HealthCheck checks if service is working correctly
func (service *Shortener) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return service.repository.Check(ctx)
}

// FormatShortURL formats url id to full url.
func (service *Shortener) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", service.config.BaseURL, urlID)
}

// // GetUrlsCreatedBy returns array of all urs that was shortened by given userID.
// // It's just a wrapper for repository.GetUsersUrls.
// func (service *Shortener) GetUrlsCreatedBy(ctx context.Context, userID string) ([]entity.ShortURL, error) {
// 	return service.repository.GetUsersUrls(ctx, userID)
// }

// ShortenBatch shortens array of urls.
// All entries of batch must contain OriginalURL.
func (service *Shortener) ShortenBatch(ctx context.Context, batch []entity.ShortURL) ([]entity.ShortURL, error) {
	// for i, URL := range batch {
	// 	urlID, err := service.generator.GenerateIDFromString(URL.OriginalURL)
	// 	if err != nil {
	// 		return nil, err
	// 	}
	// 	batch[i].ID = urlID
	// 	batch[i].CreatedByID = userID
	// }

	for i := range batch {
		id := service.Random.NewRandomString()
		// Если указать на прямую, (без конструкции id := service.Random.NewRandomString())
		// ID один и тот же!
		batch[i].ID = id // service.Random.NewRandomString()
		fmt.Println(i, batch[i].ID)
	}

	if err := service.repository.SaveBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// // GenerateNewUserID generates new user id.
// // It's just a wrapper for random.GenerateNewUserID().
// func (service *Shortener) GenerateNewUserID() string {
// 	return service.Random.GenerateNewUserID()
// }

// // DeleteUrls deletes all urls with given ids that was created by userID.
// // Implements fan-in and fan-out pattern for learning purposes.
// func (service *Shortener) DeleteUrls(ctx context.Context, ids []string, userID string) {
// 	done := make(chan struct{})
// 	defer close(done)

// 	workersCount := runtime.NumCPU()
// 	inputCh := make(chan string)
// 	entityToDelete := make([]entity.ShortURL, 0, len(ids))

// 	go func() {
// 		for _, id := range ids {
// 			inputCh <- id
// 		}

// 		close(inputCh)
// 	}()

// 	workerChs := make([]chan entity.ShortURL, 0, workersCount)
// 	for urlID := range inputCh {
// 		workerCh := make(chan entity.ShortURL)
// 		newWorker(urlID, userID, workerCh)
// 		workerChs = append(workerChs, workerCh)
// 	}

// 	for v := range fanIn(done, workerChs...) {
// 		entityToDelete = append(entityToDelete, v)
// 	}

// 	err := service.repository.DeleteUrls(ctx, entityToDelete)
// 	if err != nil {
// 		fmt.Printf("couldn't delete urls: %v\n", err)
// 	}
// }

// func (service *Shortener) GetStats(ctx context.Context) (entity.Stats, error) {
// 	usersCount, urlsCount, err := service.repository.GetUsersAndUrlsCount(ctx)
// 	if err != nil {
// 		return entity.Stats{}, err
// 	}

// 	return entity.Stats{UsersCount: usersCount, UrlsCount: urlsCount}, nil
// }

// func newWorker(urlID string, userID string, out chan entity.ShortURL) {
// 	go func() {
// 		defer func() {
// 			if x := recover(); x != nil {
// 				newWorker(urlID, userID, out)
// 				log.Printf("run time panic: %v, %v", x, out)
// 			}
// 		}()

// 		out <- entity.ShortURL{ID: urlID, CreatedByID: userID}
// 		close(out)
// 	}()
// }

// func fanIn(done <-chan struct{}, channels ...chan entity.ShortURL) chan entity.ShortURL {
// 	var wg sync.WaitGroup
// 	multiplexedStream := make(chan entity.ShortURL)

// 	multiplex := func(c <-chan entity.ShortURL) {
// 		defer wg.Done()
// 		for v := range c {
// 			select {
// 			case <-done:
// 				return
// 			case multiplexedStream <- v:
// 			}
// 		}
// 	}

// 	wg.Add(len(channels))
// 	for _, c := range channels {
// 		go multiplex(c)
// 	}

// 	go func() {
// 		wg.Wait()
// 		close(multiplexedStream)
// 	}()

// 	return multiplexedStream
// }
