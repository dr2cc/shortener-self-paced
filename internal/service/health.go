package service

import (
	"context"
	"time"
)

// HealthCheck проверяет корректность работы выбранного хранилища
func (sh *Service) HealthCheck(ctx context.Context) error {
	timeout := 5 * time.Second //nolint:gomnd
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	return sh.repository.Check(ctx)
}
