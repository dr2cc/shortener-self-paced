package jsonstore

import (
	"app/internal/domain/link"
	err_repo "app/internal/lib/repository"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

// FileRepository — репозиторий, использующий файлы для хранения.
type FileRepository struct {
	file   *os.File      // file that we will be writing to
	writer *bufio.Writer // buffered writer that will write to the file
	mutex  sync.RWMutex  // mutex that will be used to synchronize access to the file
}

// Конструктор NewFileRepository creates new file repository.
// Creates file at filePath if it doesn't exist.
// It opens a file, creates a buffered writer, and returns a pointer to a FileRepository.
func NewFileRepository(filePath string) (*FileRepository, error) {
	file, err := os.OpenFile(filePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o777) //nolint:gomnd
	if err != nil {
		return nil, err
	}

	return &FileRepository{
		mutex:  sync.RWMutex{},
		file:   file,
		writer: bufio.NewWriter(file),
	}, nil
}

// SaveBatch сохраняет несколько URL-адресов.
// Проверяет уникальность URL-адресов и сохраняет их.
func (repo *FileRepository) SaveBatch(ctx context.Context, batch []link.ExpandedURL) error {
	for _, shortURL := range batch {
		_, err := repo.FindByID(ctx, shortURL.ID)
		if err == nil {
			return err_repo.NewNotUniqueURLError(shortURL, nil)
		}
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	for _, shortURL := range batch {

		data, err := json.Marshal(shortURL)
		if err != nil {
			return err
		}

		if _, errWrite := repo.writer.Write(data); errWrite != nil {
			return errWrite
		}

		if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
			return err
		}

	}
	if err := repo.writer.Flush(); err != nil {
		return err
	}

	return nil
}

// Save проверяет уникальность URL-адреса и сохраняет его
func (repo *FileRepository) Save(ctx context.Context, shortURL link.ExpandedURL) error {
	repo.mutex.Lock() // Одна блокировка на всё
	defer repo.mutex.Unlock()

	// 1. Сначала ищем дубликат именно по OriginalURL (Iter 13)
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return err
	}
	scanner := bufio.NewScanner(repo.file)
	for scanner.Scan() {
		var entry link.ExpandedURL
		json.Unmarshal(scanner.Bytes(), &entry)
		if entry.OriginalURL == shortURL.OriginalURL {
			// Если нашли URL — возвращаем 409 и старую запись
			return err_repo.NewNotUniqueURLError(entry, nil)
		}
		if entry.ID == shortURL.ID {
			return err_repo.ErrIDCollision // Техническая коллизия ID
		}
	}

	// 2. Если всё уникально — пишем в конец
	data, _ := json.Marshal(shortURL)
	if _, err := repo.writer.Write(append(data, '\n')); err != nil {
		return err
	}
	return repo.writer.Flush()
}

// FindByID находит URL по идентификатору.
// Считывает файл строка за строкой и возвращает URL, соответствующий указанному идентификатору.
func (repo *FileRepository) FindByID(_ context.Context, id string) (link.ExpandedURL, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return link.ExpandedURL{}, err
	}

	var entry link.ExpandedURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return link.ExpandedURL{}, err
		}
		if entry.ID == id {
			return entry, nil
		}
	}

	return link.ExpandedURL{}, errors.New("can't find full url by id")
}
