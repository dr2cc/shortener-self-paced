package cache

import (
	"app/internal/domain/link"
	err_repo "app/internal/errors/repository"
	"context"
	"errors"
	"sync"
)

// InMemoryRepository — репозиторий, использующий память для хранения.
type InMemoryRepository struct {
	links map[string]link.ExpandedURL // map that will store urls
	mutex sync.RWMutex
}

// Возвращает указатель на InMemoryRepository
func NewInMemoryRepository() *InMemoryRepository {
	return &InMemoryRepository{
		links: make(map[string]link.ExpandedURL),
		mutex: sync.RWMutex{},
	}
}

// SaveBatch сохраняет несколько URL-адресов.
// Проверяет уникальность URL-адресов и сохраняет их.
func (repo *InMemoryRepository) SaveBatch(_ context.Context, batch []link.ExpandedURL) error {
	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {
		_, ok := repo.links[shortURL.ID]
		if ok {
			return err_repo.NewNotUniqueURLError(shortURL, nil)
		}
	}

	for _, shortURL := range batch {
		repo.links[shortURL.ID] = shortURL
	}

	return nil
}

// ❌ Exists проверяет существование shortURL в базе.
// ❗Возвращая ошибку (в данном случае всегда равную nil), мы выполняем контракт интерфейса:
// интерфейс ShortURL "обещает", что метод может вернуть ошибку,
// значит сервис обязан уважать этот контракт!
func (repo *InMemoryRepository) Exists(_ context.Context, shortURL string) (bool, error) {
	// Используем RLock (ReadOnly), чтобы не блокировать другие операции чтения
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	// "Comma-ok" idiom - используется в Go везде, где операция может иметь два возможных исхода, которые невозможно однозначно интерпретировать,
	// основываясь только на возвращаемом значении:
	// 1. Поиск/Извлечение: (Map, Каналы, reflect). Проверка наличия элемента или того, что канал не закрыт.
	// 2. Проверка соответствия: (Type Assertion). Проверка того, соответствует ли базовый тип интерфейса ожидаемому конкретному типу.
	// 3...

	// Здесь мы проверяем наличие ключа.
	// Если он уже есть, значит это дубль
	// ok == true
	_, ok := repo.links[shortURL]
	return ok, nil
}

// Save проверяет уникальность URL-адреса и сохраняет его
func (repo *InMemoryRepository) Save(_ context.Context, shortURL link.ExpandedURL) error {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()
	// // "Comma-ok" idiom - используется в Go везде, где операция может иметь два возможных исхода, которые невозможно однозначно интерпретировать,
	// // основываясь только на возвращаемом значении:
	// // 1. Поиск/Извлечение: (Map, Каналы, reflect). Проверка наличия элемента или того, что канал не закрыт.
	// // 2. Проверка соответствия: (Type Assertion). Проверка того, соответствует ли базовый тип интерфейса ожидаемому конкретному типу.
	// // 3...

	// // Здесь мы проверяем наличие ключа (ID).
	// // Если он уже есть, значит это дубль
	// // ok == true
	// _, ok := repo.links[shortURL.ID]

	// // В defer выше
	// //repo.mutex.RUnlock()

	// // Если выше мы уже нашли такой ключ, то вернем не nil, а
	// // &NotUniqueURLError{
	// //	Err:      err,
	// //	ShortURL: shortURL,
	// // }
	// if ok {
	// 	return err_repo.NewNotUniqueURLError(shortURL, nil)
	// }

	// repo.mutex.Lock()
	// repo.links[shortURL.ID] = shortURL
	// repo.mutex.Unlock()

	// 1. Проверка на конфликт URL (бизнес-логика Iter 13)
	for _, existing := range repo.links {
		if existing.OriginalURL == shortURL.OriginalURL {
			return err_repo.NewNotUniqueURLError(existing, nil)
		}
	}

	// 2. Проверка на коллизию ID (техническая проверка рандома)
	if _, exists := repo.links[shortURL.ID]; exists {
		return err_repo.ErrIDCollision // Специальная ошибка для повтора генерации
	}

	repo.links[shortURL.ID] = shortURL

	return nil
}

// FindByID находит URL по идентификатору.
func (repo *InMemoryRepository) FindByID(_ context.Context, id string) (link.ExpandedURL, error) {
	repo.mutex.RLock()
	url, ok := repo.links[id]
	repo.mutex.RUnlock()

	if !ok {
		return link.ExpandedURL{}, errors.New("can't find full url by id")
	}

	return url, nil
}

// Close clears map.
func (repo *InMemoryRepository) Close(_ context.Context) error {
	repo.links = make(map[string]link.ExpandedURL)
	return nil
}

// Stub function
func (repo *InMemoryRepository) Check(_ context.Context) error {
	return nil
}
