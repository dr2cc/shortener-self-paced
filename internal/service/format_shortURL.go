package service

import (
	"fmt"
)

// FormatShortURL форматирует полученный идентификатор URL
// в результирующую строку, возвращаемую запросами POST
func (sh *Service) FormatShortURL(urlID string) string {
	return fmt.Sprintf("%s/%s", sh.config.BaseURL, urlID)
}
