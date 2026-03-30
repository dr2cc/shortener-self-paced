package handler

import (
	"app/internal/domain/link"
	mock_service "app/internal/service/mocks"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

func TestController_Redirect(t *testing.T) {
	// Объявляем имитацию поведения FindURL, игнорируя контекст (контекст - это ответственность сервиса, а не контроллера)
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
					FindURL(gomock.Any(), alias). // ✅ gomock.Any()= "принимай любой контекст, мне все равно"
					Return(link.ExpandedURL{OriginalURL: "https://example.com"}, nil)
			},
		},
		{
			name:               "not found empty url",
			alias:              "notfound",
			expectedStatusCode: http.StatusBadRequest,
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
			// Создаёт мок‑объекты (сгенерированные NewMockShortURL) для интерфейсов ShortURL и Pinger.
			// Мок‑объект привязывается к ctrl и будет использоваться вместо реального сервиса.
			svc := mock_service.NewMockShortURL(ctrl)
			ping := mock_service.NewMockPinger(ctrl)
			// Вызываем функцию‑поле (структуры testTable) mockBehavior БЕЗ ctx
			// Настроиваем поведение мока svc в текущем тест-кейсе для входных данных tt.alias
			tt.mockBehavior(svc, tt.alias)
			// Создаем handler с обоими интерфейсами
			// теперь хендлер использует мок‑сервисы при своей работе
			handler := Controller{
				shortener: svc,
				health:    ping,
			}
			// Создаем тестовый роутер, без реального запуска HTTP‑сервера
			r := chi.NewRouter()
			// Регистрируем маршрут
			r.Get("/{id}", handler.Redirect)

			w := httptest.NewRecorder()                          // Имитация ResponseWriter
			req := httptest.NewRequest("GET", "/"+tt.alias, nil) // Имитация запроса

			// Имитация работы HTTP‑сервера в памяти
			r.ServeHTTP(w, req)

			// Проверяет, что HTTP‑статус‑код ответа (w.Code) совпадает с ожидаемым test.expectedStatusCode (200, 400, ...).
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			// В случае если tt.expectedStatusCode == 307 (успешный редирект). Остальные кейсы (400, 500) НЕ имеют заголовков Location
			// if защищает от паники когда заголовков Location нет
			if tt.expectedStatusCode == http.StatusTemporaryRedirect {
				assert.Equal(t, tt.url, w.Header().Get("Location"))
				assert.Equal(t, "text/html", w.Header().Get("Content-Type"))
			}
		})
	}
}
