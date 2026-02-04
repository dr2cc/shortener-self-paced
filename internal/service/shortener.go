package service

import (
	"app/internal/config"
	"app/internal/domain/link"
	"app/internal/lib/generator"
	err_repo "app/internal/lib/repository"
	storage "app/internal/repository"
	"context"
	"errors"
	"fmt"
)

// TODO: move to config if needed
const maxRetries = 3
const idLength = 8 // Оптимально для 200+ млрд комбинаций

type ShortService struct {
	// "Общение" с репозиторием, сервиса сокращения URL
	repo      storage.ShortURLRepository
	generator *generator.StringGenerator
	cfg       *config.Config // 🤷‍♂️ Нужно передавать только, что здесь нужно
	// (cfg может очень большим). Или только базовый URL или (видимо  лучше) структура в которую смапали cfg (только нужное поле)
}

func NewShortService(repo storage.ShortURLRepository, gen *generator.StringGenerator, cfg *config.Config) *ShortService {
	return &ShortService{
		repo:      repo,
		generator: gen,
		cfg:       cfg,
	}
}

// ♊Принцип:
// Handler (к примеру BatchShortenAPI) принимает запрос и передает данные в сервис.
// Service (shortener) решает, что нужно создать ссылку❗
// Link (Домен) предоставляет фабрику link.New() для сборки объекта.

// ShortenBatch мапит массив входящих данных в []link.ExpandedURL
func (sh ShortService) ShortenBatch(ctx context.Context, batch []BatchInput) ([]link.ExpandedURL, error) {
	newLinkBatch := make([]link.ExpandedURL, len(batch))

	// Обход полученного массива
	for i, URL := range batch {
		// Метод NewRandomString — это «черный ящик».
		// Сервис не знает как он работает и ему это не нужно.
		ID := sh.generator.NewRandomString(idLength)

		// Заполняем элемент массива сущностью полученной при помощи фабрики.
		newLinkBatch[i] = link.New(URL.OriginalURL, ID, link.WithCorrelationID(URL.CorrelationID))
	}

	// Попытка записи в хранилище.
	// SaveBatch реализован как транзакция и проверяет уникальность URL-адресов.
	if err := sh.repo.SaveBatch(ctx, newLinkBatch); err != nil {
		return nil, fmt.Errorf("batch save failed: %w", err)
	}

	return newLinkBatch, nil
}

func (sh *ShortService) ShortenURL(ctx context.Context, originalURL string) (link.ExpandedURL, error) {
	// Рандом может выдать уже существующий в базе ID, поэтому в месте вызова генератора (сервисном слое) нужно добавить цикл.
	// Технически это называется "Optimistic Retry Loop":
	for i := 0; i < maxRetries; i++ {
		// Получаем строку ID через метод NewRandomString.
		// Метод NewRandomString — это «черный ящик».
		// Сервису ShortURL всё равно, используется ли внутри math/rand, crypto/rand или просто вырезаются куски из UUID.
		// Он просит: «дай мне строку длиной idLength». Это и есть Abstraction Layer.
		ID := sh.generator.NewRandomString(idLength)

		// Создаем сущность через фабрику
		newLink := link.New(originalURL, ID)

		// Проверка на ошибку "Duplicate Key".
		err := sh.repo.Save(ctx, newLink)
		if err == nil {
			return newLink, nil // Успех
		}
		// Если база вернула "Duplicate Key" — продолжаем цикл.

		// iter13 Если это конфликт URL — сразу выходим и отдаем 409
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

// Функция FindURL находит в хранилище полный URL-адрес по указанному идентификатору.
// Возвращает заполненную структуру entity.ExpandedURL
func (sh ShortService) FindURL(ctx context.Context, id string) (link.ExpandedURL, error) {
	origURL, err := sh.repo.FindByID(ctx, id)
	if err != nil {
		return link.ExpandedURL{}, err //Shortener
	}
	return origURL, nil
}

// FormatShortURL форматирование полученного ID (путем конкатенации с BaseURL из cfg)
// в результирующую строку, возвращаемую запросами POST
func (sh ShortService) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", sh.cfg.BaseURL, urlID)
}
