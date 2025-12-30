package service

import (
	"app/internal/entity"
)

// По моему пониманю тут осуществляется маппинг данных уровня сервисов
// Хотя, вроде, маппинг это процесс установления соответствия между полями объекта
// из одного слоя и полями объекта другого слоя.
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
