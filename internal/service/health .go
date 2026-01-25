package service

import (
	storage "app/internal/repository"
	"context"
)

// ❌24.01.26 Переделать на прямое "общение" с pg!?
// Так нет бреда в storage.go и заглушек в двух других (а по делу в одном, json_store надо убрать)
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
	// Проверяем, реализует ли текущий репозиторий интерфейс DBHealthChecker
	// Используем динамическую проверку типа (type assertion)
	if pinger, ok := s.repo.(storage.DBHealthChecker); ok {
		return pinger.Check(ctx)
	}

	// Если это InMemory или File, которые не реализуют Ping,
	// возвращаем nil (считаем, что они всегда "здоровы")
	// или специфичную ошибку.
	return nil
}
