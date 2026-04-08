// Классическая реализация PRNG (Pseudo-Random Number Generator) с инкапсуляцией в структуру.
package generator

import (
	"math/rand/v2"
)

// Для сокращенных ссылок стандарт — это Base62 (цифры + латинские буквы в обоих регистрах).
// Он не содержит спецсимволов, которые могут «сломать» URL
const alphabet = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

// Stateless Provider ("безликий" поставщик услуг).
// Несмотря на то, что StringGenerator пуст (Stateless), наличие метода NewRandomString
// позволяет в любой момент изменить логику генерации, не переписывая основной код сервиса.
// Архитектурно это всё еще Abstraction Layer!
// Сервис сокращения ссылок по-прежнему не знает,
// как именно создается строка, он просто вызывает метод NewRandomString
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
