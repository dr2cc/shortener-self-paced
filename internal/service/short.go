package service

import (
	"app/internal/entity"
	storage "app/internal/repository"
	"context"
	"errors"
	"fmt"
)

// ❌ 01.01.2026 - должен использовать mapping? Как?
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
	// urlID, err := sh.Random.GenerateIDfromString(url)
	// if err != nil {
	// 	return entity.ExpandedURL{}, err
	// }

	// shortURL := entity.ExpandedURL{
	// 	OriginalURL: url,
	// 	ID:          urlID,
	// 	// CreatedByID: userID,
	// }

	shortURL, err := sh.mapping(url)
	if err != nil {
		return entity.ExpandedURL{}, err
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
