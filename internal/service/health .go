package service

import (
	storage "app/internal/repository"
	"context"
	"time"
)

// ❌24.01.26 Переделать на прямое "общение" с pg!?
// Так нет бреда в storage.go и заглушек в двух других (а по делу в одном, json_store надо убрать)
type HelthService struct {
	// "Общение" с репозиторием, проверки работоспособности db
	repo storage.DBHealthChecker
}

func NewHelthService(repo storage.DBHealthChecker) *HelthService {
	return &HelthService{
		repo: repo,
	}
}

// 🤷‍♂️ Другой сервис!
// HealthCheck проверяет корректность работы выбранного хранилища
func (hs HelthService) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return hs.repo.Check(ctx)
}
