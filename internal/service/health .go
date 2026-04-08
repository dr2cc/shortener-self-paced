// Здесь только обработка "заглушки" случая когда у хранилища нет поддержки Ping (In-mem, File).
package service

import (
	"context"
)

// noOpPinger — "заглушка-оптимист" для хранилищ без поддержки Ping.
type noOpPinger struct{}

// По "подсказкам" CheckHealth должны быть Ping.
// Вообще не важно! Главное не запутаться.
func (n *noOpPinger) CheckHealth(ctx context.Context) error {
	return nil // Всегда "здоров"
}
