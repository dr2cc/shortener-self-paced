package random

import (
	"encoding/binary"
	"errors"
	"hash/fnv"
	"math/big"
)

// ex. URLGenerator, поведение (метод)- GenerateIDfromString
type Stringer interface {
	GenerateIDfromString(url string) (string, error)
}

// RandomStringGenerator реализует метод GenerateIDfromString
// интерфейса Stringer
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

	result := stringFromHash(hash)
	return result, nil
}

// hashURL принимает строку и возвращает 32-битный хеш этой строки
func hashURL(url string) (uint32, error) {
	hash := fnv.New32a()
	if _, err := hash.Write([]byte(url)); err != nil {
		return 0, err
	}
	return hash.Sum32(), nil
}

// (ex. toBase62) преобразует uint32 в строку
func stringFromHash(id uint32) string {
	var i big.Int
	size := 8
	bytes := make([]byte, size)
	binary.LittleEndian.PutUint32(bytes, id)
	i.SetBytes(bytes)
	base := 62
	return i.Text(base)
}
