// Package generator is used for generating hash from string.
package generator

// Единственный метод (поведение)
// (base_64_hash_generator)generator.GenerateIDFromString
// Задача usecase связать поведение бизнес-логики и "внешний мир".
// Usecase сосредоточен на выполнении одной конкретной задачи,
// обеспечивая высокую изоляцию и строгое соблюдение принципа единственной ответственности.
// https://habr.com/ru/articles/881918/
// Видимо сейчас (15.12.2025) строго соответствует usecase только generator(?)
type URLGenerator interface {
	GenerateIDFromString(url string) (string, error)
}
