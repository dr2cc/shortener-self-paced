package service

import (
	"app/internal/entity"
)

// По моему пониманю тут осуществляется маппинг (преобразование) данных уровня сервисов.
// Если логика сервиса сфокусирована именно на создании (инстанцировании)
// сложной структуры из простых входных параметров, то ее называют ❗Factory (фабрика).
// ❌ Самостоятельный се6рвис! Не связан с ShortURL!
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
