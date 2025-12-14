// Package generator is used for generating hash from string.
package generator

// Единственный метод (поведение)
// (base_64_hash_generator)generator.GenerateIDFromString
type URLGenerator interface {
	GenerateIDFromString(url string) (string, error)
}
