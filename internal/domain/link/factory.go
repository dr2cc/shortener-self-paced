package link

// Сервис реализует инстанцирование (или сборку) доменной сущности ExpandedURL.
// Процесс включает в себя маппинг входного URL и обогащение объекта уникальным идентификатором
// с помощью внутренней логики генерации ключей.

// // 🤷‍♂️ Про factory тоже говорил. Послушать еще раз.
// // или так. Этот последний
// func New(url, id, corr string) ExpandedURL {
// 	return ExpandedURL{OriginalURL: url, ID: id, CorrelationID: corr}
// }

type Option func(*ExpandedURL)

func WithCorrelationID(id string) Option {
	return func(u *ExpandedURL) {
		u.CorrelationID = id
	}
}

func New(url, id string, opts ...Option) ExpandedURL {
	u := ExpandedURL{OriginalURL: url, ID: id}
	for _, opt := range opts {
		opt(&u)
	}
	return u
}
