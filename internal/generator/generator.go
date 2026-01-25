package generator

import (
	"math/rand"
	"time"
)

// Для сокращенных ссылок стандарт — это Base62 (цифры + латинские буквы в обоих регистрах).
// Он не содержит спецсимволов, которые могут «сломать» URL
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

type StringGenerator struct {
	rand *rand.Rand
}

func NewStringGenerator() *StringGenerator {
	// ♊В 2026 году используем новый источник рандома для безопасности
	return &StringGenerator{
		rand: rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

func (g *StringGenerator) NewRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		b[i] = alphabet[g.rand.Intn(len(alphabet))]
	}
	return string(b)
}
