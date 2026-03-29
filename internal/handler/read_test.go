package handler

import (
	"app/internal/domain/link"
	mock_service "app/internal/service/mocks"
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/mock/gomock"
)

// func TestController_Redirect(t *testing.T) {
// 	// Объявляем имитацию поведения
// 	type mockBehavior func(s *mock_service.MockShortURL, ctx context.Context, alias string)

// 	tests := []struct {
// 		name               string // description of this test case
// 		alias              string
// 		url                string
// 		mockBehavior       mockBehavior
// 		expectedStatusCode int
// 	}{
// 		{
// 			name:               "successful redirect",
// 			alias:              "abc123",
// 			url:                "https://example.com",
// 			expectedStatusCode: http.StatusTemporaryRedirect,
// 			mockBehavior: func(s *mock_service.MockShortURL, ctx context.Context, alias string) {
// 				s.EXPECT().
// 					FindURL(ctx, alias).
// 					Return(link.ExpandedURL{OriginalURL: "https://example.com"}, nil)
// 			},
// 		},
// 		{
// 			name:               "not found empty url",
// 			alias:              "notfound",
// 			expectedStatusCode: http.StatusNotFound,
// 			mockBehavior: func(s *mock_service.MockShortURL, ctx context.Context, alias string) {
// 				s.EXPECT().
// 					FindURL(ctx, alias).
// 					Return(link.ExpandedURL{OriginalURL: ""}, nil)
// 			},
// 		},
// 		{
// 			name:               "internal server error",
// 			alias:              "error",
// 			expectedStatusCode: http.StatusInternalServerError,
// 			mockBehavior: func(s *mock_service.MockShortURL, ctx context.Context, alias string) {
// 				s.EXPECT().
// 					FindURL(ctx, alias).
// 					Return(link.ExpandedURL{}, errors.New("database connection failed"))
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run(tt.name, func(t *testing.T) {
// 			ctrl := gomock.NewController(t)
// 			// Создаёт мок‑объект (сгенерированный NewMockShortURL) для интерфейса ShortURL
// 			// Мок‑объект привязывается к ctrl и будет использоваться вместо реального сервиса.
// 			svc := mock_service.NewMockShortURL(ctrl)
// 			ping := mock_service.NewMockPinger(ctrl) // Мок для PingerUseCase

// 			// Background возвращает ненулевой, пустой Context.
// 			// Она никогда не отменяется, не имеет значений и не имеет крайнего срока выполнения.
// 			// Обычно она используется в основной функции, при инициализации и тестировании
// 			ctx := context.Background()
// 			// Вызываем функцию‑поле (структуры tests) mockBehavior для настройки ожидаемого поведения,
// 			// чтобы настроить поведение мока svc (например, какие методы вызываются и что возвращают)
// 			// для входных данных ctx и tt.alias
// 			tt.mockBehavior(svc, ctx, tt.alias)

// 			// Создаем handler с обоими интерфейсами
// 			// теперь хендлер использует мок‑сервис при своей работе
// 			handler := Controller{
// 				shortener: svc,
// 				health:    ping,
// 			}

// 			// Создаем тестовый роутер (без реального запуска HTTP‑сервера) с нашим хендлером
// 			r := chi.NewRouter()
// 			r.Get("/{id}", handler.Redirect)

// 			// Создаем тестовый запрос
// 			req, err := http.NewRequestWithContext(ctx, "GET", "/"+tt.alias, nil)
// 			require.NoError(t, err)

// 			// Выполняем запрос и проверяем результат
// 			rr := httptest.NewRecorder()
// 			r.ServeHTTP(rr, req)

// 			assert.Equal(t, tt.expectedStatusCode, rr.Code)

// 			if tt.expectedStatusCode == http.StatusTemporaryRedirect {
// 				assert.Equal(t, "https://example.com", rr.Header().Get("Location"))
// 				assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
// 			}
// 		})
// 	}
// }

func TestController_Redirect(t *testing.T) {
	// Объявляем имитацию поведения, игнорируя контекст
	type mockBehavior func(s *mock_service.MockShortURL, alias string)

	tests := []struct {
		name               string
		alias              string
		url                string
		mockBehavior       mockBehavior
		expectedStatusCode int
	}{
		{
			name:               "successful redirect",
			alias:              "abc123",
			url:                "https://example.com",
			expectedStatusCode: http.StatusTemporaryRedirect,
			mockBehavior: func(s *mock_service.MockShortURL, alias string) {
				s.EXPECT().
					FindURL(gomock.Any(), alias). // ✅ gomock.Any() для контекста!
					Return(link.ExpandedURL{OriginalURL: "https://example.com"}, nil)
			},
		},
		{
			name:               "not found empty url",
			alias:              "notfound",
			expectedStatusCode: http.StatusNotFound,
			mockBehavior: func(s *mock_service.MockShortURL, alias string) {
				s.EXPECT().
					FindURL(gomock.Any(), alias).
					Return(link.ExpandedURL{OriginalURL: ""}, nil)
			},
		},
		{
			name:               "internal server error",
			alias:              "error",
			expectedStatusCode: http.StatusInternalServerError,
			mockBehavior: func(s *mock_service.MockShortURL, alias string) {
				s.EXPECT().
					FindURL(gomock.Any(), alias).
					Return(link.ExpandedURL{}, errors.New("database connection failed"))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)
			svc := mock_service.NewMockShortURL(ctrl)
			ping := mock_service.NewMockPinger(ctrl)

			handler := Controller{
				shortener: svc,
				health:    ping,
			}

			r := chi.NewRouter()
			r.Get("/{id}", handler.Redirect)

			req, err := http.NewRequestWithContext(context.Background(), "GET", "/"+tt.alias, nil)
			require.NoError(t, err)

			// ✅ Вызываем mockBehavior БЕЗ ctx - контекст игнорируется
			tt.mockBehavior(svc, tt.alias)

			rr := httptest.NewRecorder()
			r.ServeHTTP(rr, req)

			assert.Equal(t, tt.expectedStatusCode, rr.Code)

			if tt.expectedStatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.url, rr.Header().Get("Location"))
				assert.Equal(t, "text/html", rr.Header().Get("Content-Type"))
			}
		})
	}
}
