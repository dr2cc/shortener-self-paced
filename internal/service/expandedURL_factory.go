package service

import (
	"app/internal/entity"
)

// Сервис реализует инстанцирование (или сборку) доменной сущности ExpandedURL.
// Процесс включает в себя маппинг входного URL и обогащение объекта уникальным идентификатором
// с помощью внутренней логики генерации ключей.

// По моему пониманю тут осуществляется маппинг (преобразование) данных уровня сервисов.
// Если логика сервиса сфокусирована именно на создании (инстанцировании)
// сложной структуры из простых входных параметров, то ее называют ❗Factory (фабрика).
// ❌ Самостоятельный сервис! Не связан с ShortURL!
// Тут создаем экземпляр entity.ExpandedURL{} и заполняем в нем поля
// OriginalURL, ID
func (sh *Service) mapping(url string) (entity.ExpandedURL, error) {
	urlID, err := sh.Random.GenerateIDfromString(url)
	if err != nil {
		return entity.ExpandedURL{}, err
	}

	return entity.ExpandedURL{
		OriginalURL: url,
		ID:          urlID,
	}, nil
}
