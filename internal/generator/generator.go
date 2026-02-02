// Классическая реализация PRNG (Pseudo-Random Number Generator) с инкапсуляцией в структуру.
package generator

import (
	"math/rand"
	"time"
)

// Для сокращенных ссылок стандарт — это Base62 (цифры + латинские буквы в обоих регистрах).
// Он не содержит спецсимволов, которые могут «сломать» URL
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Stateful Factory (фабрика с состоянием). Она хранит состояние генератора внутри себя.
type StringGenerator struct {
	rand *rand.Rand
}

func NewStringGenerator() *StringGenerator {
	// ♊В 2026 году используем новый источник рандома
	return &StringGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
		// rand.NewSource (Источник/Seed): Детерминированный алгоритм. Передавая UnixNano,
		// делаем "засев" (Seeding), чтобы последовательность не повторялась при каждом запуске сервиса.
	}
}

func (g *StringGenerator) NewRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[g.rand.Intn(len(alphabet))]
		// g.rand.Intn (Метод распределения): Получение случайного индекса.
		// В math/rand используется равномерное распределение, что важно для минимизации коллизий (повторов).
	}
	return string(b)
}
