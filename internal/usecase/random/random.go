package random

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/big"
)

// TODO: move to config if needed
const keyLength = 8

// ex. URLGenerator,поведение (метод)- (base_64_hash_generator)generator.GenerateIDFromString
type Stringer interface {
	GenerateIDfromString(url string) (string, error)
}

// RandomStringGenerator реализует метод GenerateIDFromString
// интерфейса (generator)generator.URLGenerator
// и может реализовать более простой NewRandomString
type RandomStringGenerator struct{}

// GenerateIDfromString создает ID (shortURL) из url.
func (RandomStringGenerator) GenerateIDfromString(str string) (string, error) {
	if str == "" {
		return "", errors.New("empty string to generate id from")
	}

	hash, err := hashURL(str)
	if err != nil {
		return "", err
	}

	result := toBase62(hash)
	return result, nil
}

// hashURL takes a string, and returns a 32-bit hash of that string.
func hashURL(url string) (uint32, error) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(url)); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}

// toBase62 converts a 32-bit integer to a base 62 string.
func toBase62(id uint32) string {
	var i big.Int
	size := 8
	bytes := make([]byte, size)
	binary.LittleEndian.PutUint32(bytes, id)
	i.SetBytes(bytes)
	base := 62
	return i.Text(base)
}

// // NewRandomString generates random string with given size.
// func (r RandomStringGenerator) NewRandomString() string {
// 	rnd := rand.New(rand.NewSource(time.Now().UnixNano()))

// 	chars := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZ" +
// 		"abcdefghijklmnopqrstuvwxyz" +
// 		"0123456789")

// 	b := make([]rune, keyLength)
// 	for i := range b {
// 		b[i] = chars[rnd.Intn(len(chars))]
// 	}

// 	return string(b)
// }
