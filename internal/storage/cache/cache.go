package cache

import (
	"app/internal/entity"
	"app/internal/storage"
	"context"
	"errors"
	"sync"
)

// InMemoryRepository is repository that uses memory for storage.
type InMemoryRepository struct {
	storage map[string]entity.ExpandedURL // map that will store urls
	mutex   sync.RWMutex
}

// Возвращает указатель на InMemoryRepository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]entity.ExpandedURL),
		mutex:   sync.RWMutex{},
	}
}

// SaveBatch saves multiple urls.
// Checks if the urls are unique and then saving them.
func (repo *InMemoryRepository) SaveBatch(_ context.Context, batch []entity.ExpandedURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		_, ok := repo.storage[shortURL.ID]
		if ok {
			return storage.NewNotUniqueURLError(shortURL, nil)
		}
	}

	for _, shortURL := range batch {
		repo.storage[shortURL.ID] = shortURL
	}

	return nil
}

// Save checks if the url is unique and then saving it to the memory.
func (repo *InMemoryRepository) Save(_ context.Context, shortURL entity.ExpandedURL) error {
	repo.mutex.RLock()
	// "Comma-ok" idiom - используется в Go везде, где операция может иметь два возможных исхода, которые невозможно однозначно интерпретировать,
	// основываясь только на возвращаемом значении:
	// 1. Поиск/Извлечение: (Map, Каналы, reflect). Проверка наличия элемента или того, что канал не закрыт.
	// 2. Проверка соответствия: (Type Assertion). Проверка того, соответствует ли базовый тип интерфейса ожидаемому конкретному типу.
	// 3...

	// Здесь мы проверяем наличие ключа (ID).
	// Если он уже есть, значит это дубль
	// ok == true
	_, ok := repo.storage[shortURL.ID]
	repo.mutex.RUnlock()

	// Если выше мы уже нашли такой ключ, то вернем не nil, а
	// &NotUniqueURLError{
	//	Err:      err,
	//	ShortURL: shortURL,
	// }
	if ok {
		return storage.NewNotUniqueURLError(shortURL, nil)
	}

	repo.mutex.Lock()
	repo.storage[shortURL.ID] = shortURL
	repo.mutex.Unlock()

	return nil
}

// GetByID gets the url by id.
func (repo *InMemoryRepository) FindByID(_ context.Context, id string) (entity.ExpandedURL, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return entity.ExpandedURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

// Close clears map.
func (repo *InMemoryRepository) Close(_ context.Context) error {
	repo.storage = make(map[string]entity.ExpandedURL)
	return nil
}

// Stub function
func (repo *InMemoryRepository) Check(_ context.Context) error {
	return nil
}
