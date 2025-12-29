package service

import (
	"app/internal/entity"
	"context"
)

// Функция FindURL находит в хранилище полный URL-адрес по указанному идентификатору.
// Возвращает заполненную структуру entity.ExpandedURL
func (sh *Service) FindURL(ctx context.Context, id string) (entity.ExpandedURL, error) {
	origURL, err := sh.repository.FindByID(ctx, id)
	if err != nil {
		return entity.ExpandedURL{}, err //Shortener
	}
	return origURL, nil
}
