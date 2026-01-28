package service

import (
	storage "app/internal/repository"
	"context"
)

type HelthService struct {
	// "Общение" с репозиторием, проверки работоспособности db
	repo storage.ShortURLRepository
}

func NewHelthService(repo storage.ShortURLRepository) *HelthService {
	return &HelthService{
		repo: repo,
	}
}

// HealthCheck проверяет корректность работы выбранного хранилища
func (s *HelthService) CheckHealth(ctx context.Context) error {
	// Проверяем, реализует ли текущий репозиторий интерфейс Pinger
	// Используем динамическую проверку типа (type assertion)
	if pinger, ok := s.repo.(storage.Pinger); ok {
		return pinger.CheckHealth(ctx)
	}

	// Если это InMemory или File, которые не реализуют Ping,
	// возвращаем nil (считаем, что они всегда "здоровы")
	// или специфичную ошибку.
	return nil
}
