package service

import (
	"context"
)

// type HelthService struct {
// 	// "Общение" с репозиторием, проверки работоспособности db
// 	repo storage.ShortURLRepository
// }

// noOpPinger — "заглушка-оптимист" для хранилищ без поддержки Ping (In-mem, File).
type noOpPinger struct{}

// По "подсказкам" все CheckHealth должны быть Ping.
// Вообще не важно! Главное не запутаться.
func (n *noOpPinger) CheckHealth(ctx context.Context) error {
	return nil // Всегда "здоров"
}

// // CheckHealth проверяет корректность работы выбранного хранилища
// func (s *HelthService) Ping(ctx context.Context) error {
// 	// Проверяем, реализует ли текущий репозиторий интерфейс Pinger
// 	// Используем динамическую проверку типа (type assertion)
// 	if pinger, ok := s.repo.(storage.Pinger); ok {
// 		return pinger.CheckHealth(ctx)
// 	}

// 	// Если это InMemory или File, которые не реализуют Ping,
// 	// возвращаем nil (считаем, что они всегда "здоровы")
// 	// или специфичную ошибку.
// 	return nil
// }
