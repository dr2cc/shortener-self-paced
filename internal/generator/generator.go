// Классическая реализация PRNG (Pseudo-Random Number Generator) с инкапсуляцией в структуру.
package generator

import (
	"math/rand/v2"
)

// Для сокращенных ссылок стандарт — это Base62 (цифры + латинские буквы в обоих регистрах).
// Он не содержит спецсимволов, которые могут «сломать» URL
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Stateful Factory (фабрика с состоянием). Она хранит состояние генератора внутри себя.
type StringGenerator struct{}

func NewStringGenerator() *StringGenerator {
	return &StringGenerator{}
}

func (g *StringGenerator) NewRandomString(length int) string {
	b := make([]byte, length)
	for i := range b {
		// rand.IntN в v2 потокобезопасен и работает быстрее
		b[i] = alphabet[rand.IntN(len(alphabet))]
	}
	return string(b)
}
