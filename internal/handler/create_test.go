package handler

import (
	"app/internal/domain/link"
	mock_service "app/internal/service/mocks"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestController_ShortenText(t *testing.T) {
	// 1️⃣Arrange
	type mockBehavior func(s *mock_service.MockShortURL, url string, alias string)

	tests := []struct {
		name               string
		url                string
		mockBehavior       mockBehavior
		expectedAlias      string
		expectedStatusCode int
	}{
		{
			name:               "successful url shortening",
			url:                "https://example.com",
			expectedStatusCode: http.StatusCreated,
			expectedAlias:      "abc123",
			mockBehavior: func(s *mock_service.MockShortURL, url string, alias string) {
				gomock.InOrder( // Устанавливает порядок вызовов. Не строго обязателен.
					// 1. Ожидаем вызов сокращения (вернет ID)
					s.EXPECT().
						ShortenURL(gomock.Any(), url).
						Return(link.ExpandedURL{OriginalURL: url, ID: alias}, nil),
					// 2. СРАЗУ ЖЕ ожидаем вызов форматирования (вернет полный URL)
					s.EXPECT().FormatShortURL(alias).
						Return(url),
				)
			},
		},
		{
			name:               "empty URL",
			url:                "",
			expectedStatusCode: http.StatusBadRequest,
			mockBehavior: func(s *mock_service.MockShortURL, url string, alias string) {
				// Ничего не пишем!
				// Мы НЕ ожидаем никаких вызовов сервиса, так как хендлер должен отсечь пустой URL сразу.
				//  Отсутствие лишних вызовов (Mocks)
				//
				// Поскольку используется gomock, проверка того, что методы сервиса НЕ вызывались при плохих входных данных
				// (как в этом тесте с пустым URL) — это тоже часть проверки поведения.
				// Если mockBehavior оставлен пустым для ошибки 400, gomock сам проверит, что лишних вызовов не было.
			},
		},
	}
	// 2️⃣Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)                // gomock сontroller
			svc := mock_service.NewMockShortURL(ctrl)      // мок сервиса
			tt.mockBehavior(svc, tt.url, tt.expectedAlias) // настройка мока
			handler := Controller{                         // хендлер с моком "внутри"
				shortener: svc,
			}
			r := chi.NewRouter()             // тестовый маршрутизатор
			r.Post("/", handler.ShortenText) // маршрут

			w := httptest.NewRecorder() // Имитация ResponseWriter
			// В httptest.NewRequest (Имитация запроса), target (второй аргумент) всегда "/..." (путь), он не может быть ""
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.url)) // тестируемый URL передаем в тело

			r.ServeHTTP(w, req) // Имитация работы HTTP‑сервера в памяти
			// 3️⃣Assert
			assert.Equal(t, tt.expectedStatusCode, w.Code)

			// // Не проверял!!!
			// // Сравниваем полученную строку с ожидаемой
			// assert.Equal(t, tt.expectedShortURL, string(resBody))
			// // Заголовки (Headers)
			// assert.Equal(t, "text/plain; charset=utf-8", w.Header().Get("Content-Type"))
			// // Длина ответа
			// //Если ID генерируется случайно (и ты не мокаешь его жестко), можно проверить хотя бы непустоту ответа или его длину.
			// assert.NotEmpty(t, string(resBody))
		})
	}
}
