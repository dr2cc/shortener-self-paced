package jsonstore

import (
	"app/internal/entity"
	"app/internal/storage"
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
func (repo *FileRepository) SaveBatch(ctx context.Context, batch []entity.ExpandedURL) error {
	for _, shortURL := range batch {
		_, err := repo.FindByID(ctx, shortURL.ID)
		if err == nil {
			return storage.NewNotUniqueURLError(shortURL, nil)
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
func (repo *FileRepository) Save(ctx context.Context, shortURL entity.ExpandedURL) error {
	_, err := repo.FindByID(ctx, shortURL.ID)
	if err == nil {
		return storage.NewNotUniqueURLError(shortURL, nil)
	}

	data, err := json.Marshal(shortURL)
	if err != nil {
		return err
	}

	repo.mutex.Lock()
	defer repo.mutex.Unlock()

	if _, errWrite := repo.writer.Write(data); errWrite != nil {
		return errWrite
	}

	if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
		return errWriteByte
	}

	if errFlush := repo.writer.Flush(); errFlush != nil {
		return errFlush
	}

	return nil
}

// FindByID находит URL по идентификатору.
// Считывает файл строка за строкой и возвращает URL, соответствующий указанному идентификатору.
func (repo *FileRepository) FindByID(_ context.Context, id string) (entity.ExpandedURL, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return entity.ExpandedURL{}, err
	}

	var entry entity.ExpandedURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return entity.ExpandedURL{}, err
		}
		if entry.ID == id {
			return entry, nil
		}
	}

	return entity.ExpandedURL{}, errors.New("can't find full url by id")
}

// Close closes file.
func (repo *FileRepository) Close(_ context.Context) error {
	return repo.file.Close()
}

// Check checks if file is ok.
func (repo *FileRepository) Check(_ context.Context) error {
	_, err := repo.file.Stat()
	return err
}
