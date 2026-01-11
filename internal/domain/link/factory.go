package link

// Сервис реализует инстанцирование (или сборку) доменной сущности ExpandedURL.
// Процесс включает в себя маппинг входного URL и обогащение объекта уникальным идентификатором
// с помощью внутренней логики генерации ключей.

// По моему пониманю тут осуществляется маппинг (преобразование) данных уровня сервисов.
// Если логика сервиса сфокусирована именно на создании (инстанцировании)
// сложной структуры из простых входных параметров, то ее называют ❗Factory (фабрика).
// ❌ Самостоятельный сервис! Не связан с ShortURL!
// Тут создаем экземпляр entity.ExpandedURL{} и заполняем в нем поля
// OriginalURL, ID
func Mapping(url string) (ExpandedURL, error) {
	urlID, err := GenerateIDfromString(url)
	if err != nil {
		return ExpandedURL{}, err
	}

	return ExpandedURL{
		OriginalURL: url,
		ID:          urlID,
	}, nil
}

// // Gemini предложил такой
// func New(url string, gen IDGenerator) ExpandedURL {
//     return ExpandedURL{
//         OriginalURL: url,
//         ID:          gen.NewRandomString(),
//     }
// }

// // или так. Этот последний
// func New(url, id string) ExpandedURL {
// 	return ExpandedURL{originalURL: url, id: id}
// }
