package service

import (
	"app/internal/config"
	"app/internal/entity"
	err_repo "app/internal/errors/repository"
	err_service "app/internal/errors/service"
	storage "app/internal/repository"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
)

type ShortService struct {
	// "Общение" с репозиторием, сервиса сокращения URL
	repo storage.ShortURL
	cfg  *config.Config
}

func NewShortService(repo storage.ShortURL, cfg *config.Config) *ShortService {
	return &ShortService{
		repo: repo,
		cfg:  cfg,
	}
}

// GenerateIDfromString создает ID (shortURL) из url.
func (ShortService) GenerateIDfromString(str string) (string, error) {
	if str == "" {
		return "", errors.New("empty string to generate id from")
	}

	hash, err := hashURL(str)
	if err != nil {
		return "", err
	}

	result := stringFromHash(hash)
	return result, nil
}

// FormatShortURL форматирование полученного ID (путем конкатенации с BaseURL из cfg)
// в результирующую строку, возвращаемую запросами POST
func (sh ShortService) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", sh.cfg.BaseURL, urlID)
}

// ❌ Уже на входе ошибка слоев! Мы получаем нашу готовую модель данных.
// Получается ее готовит слой handlers!!!
// 01.01.2026 - должен использовать mapping
// и получать на вход необработанную структуру из запроса.
//
// ShortenBatch мапит массив входящих данных в []entity.ExpandedURL
// Все записи пакета должны содержать OriginalURL(?)
func (sh ShortService) ShortenBatch(ctx context.Context, batch []entity.ExpandedURL) ([]entity.ExpandedURL, error) {
	for i, URL := range batch {
		urlID, err := sh.GenerateIDfromString(URL.OriginalURL)
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

// Shorten сокращает полный URL и возвращает заполненную структуру ExpandedURL
func (sh ShortService) Shorten(ctx context.Context, url string) (entity.ExpandedURL, error) {
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
	var notUniqueErr *err_repo.NotUniqueURLError
	// 🔦 func errors.As(err error, target any) bool
	// As находит первую ошибку в дереве err, соответствующую target, и, если она найдена,
	// устанавливает target равным этому значению ошибки и возвращает true.
	// В противном случае возвращает false.
	// Дерево состоит из самого err, за которым следуют ошибки,
	// полученные путём многократного вызова его метода Unwrap() error или Unwrap() []error.
	// Когда err оборачивает несколько ошибок, As проверяет err, а затем выполняет обход в глубину его дочерних элементов.
	if errors.As(err, &notUniqueErr) {
		// Если запись не уникальна, возвращаем:
		return shortURL, err_service.NewShorteningError(shortURL, err)
	}

	if err != nil {
		return entity.ExpandedURL{}, err
	}

	return shortURL, nil
}

// hashURL принимает строку и возвращает 32-битный хеш этой строки
func hashURL(url string) (uint32, error) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(url)); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}

// (ex. toBase62) преобразует uint32 в строку
func stringFromHash(id uint32) string {
	var i big.Int
	size := 8
	bytes := make([]byte, size)
	binary.LittleEndian.PutUint32(bytes, id)
	i.SetBytes(bytes)
	base := 62
	return i.Text(base)
}
