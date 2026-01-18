package service

import (
	"app/internal/config"
	"app/internal/domain/link"
	err_repo "app/internal/errors/repository"
	"app/internal/generator"
	storage "app/internal/repository"
	"context"
	"encoding/binary"
	"errors"
	"fmt"
	"hash/fnv"
	"math/big"
	"time"
)

type ShortService struct {
	// "Общение" с репозиторием, сервиса сокращения URL
	repo      storage.ShortURL
	generator *generator.StringGenerator
	cfg       *config.Config // 🤷‍♂️ Нужно передавать только, что здесь нужно
	// (cfg может очень большим). Или только базовый URL или (видимо  лучше) структура в которую смапали cfg (только нужное поле)
}

func NewShortService(repo storage.ShortURL, gen *generator.StringGenerator, cfg *config.Config) *ShortService {
	return &ShortService{
		repo:      repo,
		generator: gen,
		cfg:       cfg,
	}
}

// Функция FindURL находит в хранилище полный URL-адрес по указанному идентификатору.
// Возвращает заполненную структуру entity.ExpandedURL
func (sh ShortService) FindURL(ctx context.Context, id string) (link.ExpandedURL, error) {
	origURL, err := sh.repo.FindByID(ctx, id)
	if err != nil {
		return link.ExpandedURL{}, err //Shortener
	}
	return origURL, nil
}

// 🤷‍♂️ Другой сервис!
// HealthCheck проверяет корректность работы выбранного хранилища
func (sh ShortService) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return sh.repo.Check(ctx)
}

// FormatShortURL форматирование полученного ID (путем конкатенации с BaseURL из cfg)
// в результирующую строку, возвращаемую запросами POST
func (sh ShortService) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", sh.cfg.BaseURL, urlID)
}

// ❌ ShortenBatch - уже на входе ошибка слоев! Мы получаем нашу готовую модель данных.
// Получается ее готовит слой handlers!!!
// 01.01.2026 - должен использовать mapping
// и получать на вход необработанную структуру из запроса.
//
// ShortenBatch мапит массив входящих данных в []entity.ExpandedURL
// Все записи пакета должны содержать OriginalURL(?)
func (sh ShortService) ShortenBatch(ctx context.Context, batch []link.ExpandedURL) ([]link.ExpandedURL, error) {
	for i, URL := range batch {
		urlID, err := GenerateIDfromString(URL.OriginalURL)
		if err != nil {
			return nil, err
		}
		batch[i].ID = urlID
		//batch[i].CreatedByID = userID
	}

	if err := sh.repo.SaveBatch(ctx, batch); err != nil {
		return nil, err
	}

	return batch, nil
}

// // Shorten сокращает полный URL и возвращает заполненную структуру ExpandedURL
// func (sh ShortService) Shorten(ctx context.Context, url string) (entity.ExpandedURL, error) {
// 	shortURL, err := Mapping(url)
// 	if err != nil {
// 		return entity.ExpandedURL{}, err
// 	}

// 	// Пробуем записать в хранилище, с проверкой уникальности (iter13)
// 	err = sh.repo.Save(ctx, shortURL)

// 	// "Реакция" на уникальность / не уникальность
// 	var notUniqueErr *err_repo.NotUniqueURLError
// 	// 🔦 func errors.As(err error, target any) bool
// 	// As находит первую ошибку в дереве err, соответствующую target, и, если она найдена,
// 	// устанавливает target равным этому значению ошибки и возвращает true.
// 	// В противном случае возвращает false.
// 	// Дерево состоит из самого err, за которым следуют ошибки,
// 	// полученные путём многократного вызова его метода Unwrap() error или Unwrap() []error.
// 	// Когда err оборачивает несколько ошибок, As проверяет err, а затем выполняет обход в глубину его дочерних элементов.
// 	if errors.As(err, &notUniqueErr) {
// 		// Если запись не уникальна, возвращаем:
// 		return shortURL, err_service.NewShorteningError(shortURL, err)
// 	}

// 	if err != nil {
// 		return entity.ExpandedURL{}, err
// 	}

// 	return shortURL, nil
// }

func (sh *ShortService) CreateShortURL(ctx context.Context, originalURL string) (link.ExpandedURL, error) {
	const maxRetries = 3
	const idLength = 8 // Оптимально для 200+ млрд комбинаций

	for i := 0; i < maxRetries; i++ {
		// 1. Генерируем случайный ID
		newID := sh.generator.NewRandomString(idLength)

		// Создаем сущность через фабрику
		newLink := link.New(originalURL, newID)

		err := sh.repo.Save(ctx, newLink)
		if err == nil {
			return newLink, nil // Успех
		}

		// Если это конфликт URL — сразу выходим и отдаем 409
		var notUniqueErr *err_repo.NotUniqueURLError
		if errors.As(err, &notUniqueErr) {
			return notUniqueErr.ShortURL, err
		}

		// Если это коллизия ID — идем на следующую итерацию цикла (i++)
		if errors.Is(err, err_repo.ErrIDCollision) {
			continue
		}

		// "Защитное программирование".
		// Эта ошибка на тот случай, если все наши if не сработали.
		// При in-memory хранилище мы сюда не попадем.
		// Но если база будет Postgres — этот код спасет нас, например от краша при сбое сети.
		return link.ExpandedURL{}, err
	}
	// Эта ошибка возникнет если сервису не удалось создать уникальный ID после максимально допустимого количества попыток
	return link.ExpandedURL{}, err_repo.ErrGenerationFailed
}

// GenerateIDfromString создает ID (shortURL) из url.
func GenerateIDfromString(str string) (string, error) {
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
