package storage

import (
	"app/internal/entity"
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"sync"
)

// Структура такая. Тип FileRepository и 10 его методов

// FileRepository is repository that uses files for storage.
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

func (repo *FileRepository) SaveBatch(ctx context.Context, batch []entity.ShortURL) error {
	for _, shortURL := range batch {
		_, err := repo.GetByID(ctx, shortURL.ID)
		if err == nil {
			return NewNotUniqueURLError(shortURL, nil)
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

// Save checks if the url is unique and then saving it to the file.
func (repo *FileRepository) Save(ctx context.Context, shortURL entity.ShortURL) error {
	_, err := repo.GetByID(ctx, shortURL.ID)
	if err == nil {
		return NewNotUniqueURLError(shortURL, nil)
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

// GetByID gets url by id.
// Reads the file line by line and returns url that matches given id.
func (repo *FileRepository) GetByID(_ context.Context, id string) (entity.ShortURL, error) {
	repo.mutex.RLock()
	defer repo.mutex.RUnlock()

	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return entity.ShortURL{}, err
	}

	var entry entity.ShortURL

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return entity.ShortURL{}, err
		}
		if entry.ID == id {
			return entry, nil
		}
	}

	return entity.ShortURL{}, errors.New("can't find full url by id")
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

// readFileToMap reads the file and returns a map of all the urls in the file.
func (repo *FileRepository) readFileToMap() (map[string]entity.ShortURL, error) {
	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
		return nil, err
	}
	var entry entity.ShortURL
	existingURLs := make(map[string]entity.ShortURL)

	scanner := bufio.NewScanner(repo.file)

	for scanner.Scan() {
		line := scanner.Bytes()
		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
			return nil, err
		}
		existingURLs[entry.ID] = entry
	}
	return existingURLs, nil
}

// writeMapToFile writes the map to the file.
func (repo *FileRepository) writeMapToFile(existingURLs map[string]entity.ShortURL) error {
	if err := repo.file.Truncate(0); err != nil {
		return err
	}
	if _, err := repo.file.Seek(0, 0); err != nil {
		return err
	}

	for _, url := range existingURLs {

		data, err := json.Marshal(url)
		if err != nil {
			return err
		}

		if _, errWrite := repo.writer.Write(data); errWrite != nil {
			return errWrite
		}

		if errWriteByte := repo.writer.WriteByte('\n'); errWriteByte != nil {
			return errWriteByte
		}

	}
	if err := repo.writer.Flush(); err != nil {
		return err
	}
	return nil
}

// // GetUsersUrls reads the file line by line and returning all the urls
// // that were created by user with id userID.
// func (repo *FileRepository) GetUsersUrls(_ context.Context, userID string) ([]entity.ShortURL, error) {
// 	repo.mutex.RLock()
// 	defer repo.mutex.RUnlock()

// 	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
// 		return nil, err
// 	}

// 	var entry entity.ShortURL
// 	var URLs []entity.ShortURL

// 	scanner := bufio.NewScanner(repo.file)

// 	count := 0
// 	for scanner.Scan() {
// 		line := scanner.Bytes()
// 		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
// 			return nil, err
// 		}
// 		count += 1
// 		fmt.Println(count)

// 		if entry.CreatedByID == userID {
// 			URLs = append(URLs, entry)
// 		}
// 	}

// 	return URLs, nil
// }

// // DeleteUrls deletes all given urls.
// func (repo *FileRepository) DeleteUrls(_ context.Context, urls []entity.ShortURL) error {
// 	repo.mutex.Lock()
// 	defer repo.mutex.Unlock()

// 	existingURLs, err := repo.readFileToMap()
// 	if err != nil {
// 		return err
// 	}

// 	// mark deleted urls in memory
// 	now := time.Now()
// 	for _, urlToDelete := range urls {
// 		foundURL, ok := existingURLs[urlToDelete.ID]
// 		if ok && foundURL.CreatedByID == urlToDelete.CreatedByID {
// 			foundURL.DeletedAt = now
// 			existingURLs[urlToDelete.ID] = foundURL
// 		}
// 	}

// 	// write back in memory map to file
// 	err = repo.writeMapToFile(existingURLs)
// 	if err != nil {
// 		return err
// 	}

// 	return nil
// }

// func (repo *FileRepository) GetUsersAndUrlsCount(_ context.Context) (int, int, error) {
// 	if _, err := repo.file.Seek(0, io.SeekStart); err != nil {
// 		return 0, 0, err
// 	}

// 	uniqueUsersIds := make(map[string]bool)
// 	urlsCount := 0

// 	var entry entity.ShortURL
// 	scanner := bufio.NewScanner(repo.file)

// 	for scanner.Scan() {
// 		line := scanner.Bytes()
// 		if err := json.NewDecoder(bytes.NewReader(line)).Decode(&entry); err != nil {
// 			return 0, 0, err
// 		}
// 		urlsCount++
// 		uniqueUsersIds[entry.CreatedByID] = true
// 	}

// 	return len(uniqueUsersIds), urlsCount, nil
// }
