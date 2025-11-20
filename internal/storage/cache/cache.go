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
	storage map[string]entity.ShortURL // map that will store urls
	mutex   sync.RWMutex               // read-write mutex that will be used to synchronize access to the storage map
}

// NewInMemoryRepository creates a new InMemoryRepository and returns a pointer to it.
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		storage: make(map[string]entity.ShortURL),
		mutex:   sync.RWMutex{},
	}
}

// SaveBatch saves multiple urls.
// Checks if the urls are unique and then saving them.
func (repo *InMemoryRepository) SaveBatch(_ context.Context, batch []entity.ShortURL) error {
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
func (repo *InMemoryRepository) Save(_ context.Context, shortURL entity.ShortURL) error {
	repo.mutex.RLock()
	_, ok := repo.storage[shortURL.ID]
	repo.mutex.RUnlock()

	if ok {
		return storage.NewNotUniqueURLError(shortURL, nil)
	}

	repo.mutex.Lock()
	repo.storage[shortURL.ID] = shortURL
	repo.mutex.Unlock()

	return nil
}

// GetByID gets the url by id.
func (repo *InMemoryRepository) FindByID(_ context.Context, id string) (entity.ShortURL, error) {
	repo.mutex.RLock()
	url, ok := repo.storage[id]
	repo.mutex.RUnlock()

	if !ok {
		return entity.ShortURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

// Close clears map.
func (repo *InMemoryRepository) Close(_ context.Context) error {
	repo.storage = make(map[string]entity.ShortURL)
	return nil
}

// Stub function
func (repo *InMemoryRepository) Check(_ context.Context) error {
	return nil
}

// // GetUsersUrls gets all the urls that were created by the user with the given id.
// func (repo *InMemoryRepository) GetUsersUrls(_ context.Context, userID string) ([]entity.ShortURL, error) {
// 	repo.mutex.RLock()
// 	var URLs []entity.ShortURL
// 	for _, URL := range repo.storage {
// 		if URL.CreatedByID == userID {
// 			URLs = append(URLs, URL)
// 		}
// 	}
// 	repo.mutex.RUnlock()
// 	return URLs, nil
// }

// // DeleteUrls deletes all given urls.
// func (repo *InMemoryRepository) DeleteUrls(_ context.Context, urls []entity.ShortURL) error {
// 	repo.mutex.Lock()
// 	defer repo.mutex.Unlock()

// 	now := time.Now()
// 	for _, urlToDelete := range urls {
// 		foundURL, ok := repo.storage[urlToDelete.ID]
// 		if ok && foundURL.CreatedByID == urlToDelete.CreatedByID {
// 			foundURL.DeletedAt = now
// 			repo.storage[urlToDelete.ID] = foundURL
// 		}
// 	}

// 	return nil
// }

// func (repo *InMemoryRepository) GetUsersAndUrlsCount(_ context.Context) (int, int, error) {
// 	uniqueUsersIds := make(map[string]bool)

// 	repo.mutex.RLock()
// 	defer repo.mutex.RUnlock()

// 	for _, shortURL := range repo.storage {
// 		uniqueUsersIds[shortURL.CreatedByID] = true
// 	}

// 	return len(uniqueUsersIds), len(repo.storage), nil
// }
