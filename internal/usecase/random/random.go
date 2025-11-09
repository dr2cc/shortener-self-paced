package random

import (
	"math/rand"
	"time"
)

// TODO: move to config if needed
const keyLength = 6

type RandomGenerator struct{}

// NewRandomString generates random string with given size.
func (g *RandomGenerator) NewRandomString() string {
	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
		"abcdefghijklmnopqrstuvwxyz" +
		"0123456789")

	b := make([]rune, keyLength)
	for i := range b {
		b[i] = chars[rnd.Intn(len(chars))]
	}

	return string(b)
}
