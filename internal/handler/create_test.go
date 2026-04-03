package handler

import (
	"app/internal/domain/link"
	err_repo "app/internal/lib/repository"
	mock_service "app/internal/service/mocks"
	"bytes"
	"compress/gzip"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

const (
	testBaseURL   = "http://localhost:8080"
	expectedAlias = "abc123"
)

func TestController_ShortenText(t *testing.T) {
	// 🔸Arrange
	type mockBehavior func(s *mock_service.MockShortURL, url string)

	tests := []struct {
		name               string
		url                string
		expectedStatusCode int
		mockBehavior       mockBehavior
	}{
		{
			name:               "successful url shortening",
			url:                "https://example.com",
			expectedStatusCode: http.StatusCreated,
			mockBehavior: func(s *mock_service.MockShortURL, url string) {
				gomock.InOrder( // устанавливает порядок вызовов. Метод InOrder не обязателен.
					// 1️⃣ При обращении к объекту s мы будем ОЖИДАТЬ().
					s.EXPECT().
						// 2️⃣ что вызов метода (структуры MockShortURLMockRecorder) ShortenURL(gomock.Any(), url)
						ShortenURL(gomock.Any(), url).
						// 3️⃣ тестируемому коду Вернет(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, nil)
						// (имитацию ответа от Controller.shortener.ShortenURL(...) (link.ExpandedURL, error))
						Return(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, nil),
					// 2. СРАЗУ ЖЕ ожидаем вызов форматирования (вернет полный URL)
					// Следующим мы ОЖИДАЕМ().FormatShortURL()
					s.EXPECT().FormatShortURL(expectedAlias).
						Return(testBaseURL+"/"+expectedAlias),
				)
			},
		},
		{
			name:               "empty URL",
			url:                "",
			expectedStatusCode: http.StatusBadRequest,
			mockBehavior: func(s *mock_service.MockShortURL, url string) {
				// Ничего не пишем!
				// Мы НЕ ожидаем никаких вызовов сервиса, так как хендлер должен отсечь пустой URL сразу.
				//
				// "Немного" 🔸🔸🔸Assert
				// Проверка отсутствие лишних вызовов (проверка того, что методы сервиса НЕ вызывались при плохих входных данных
				// (пустой URL) — это тоже часть проверки поведения).
				// Если mockBehavior оставлен пустым для ошибки 400, gomock сам проверит, что лишних вызовов не было.
			},
		},
		{
			name:               "conflict - url already exists",
			url:                "https://example.com",
			expectedStatusCode: http.StatusConflict,
			mockBehavior: func(s *mock_service.MockShortURL, url string) {
				// gomock.InOrder() здесь не используем- работает!
				// В случае конфликта ShortenURL здесь возвращает заполненную структуру и нашу специфическую ошибку
				s.EXPECT().ShortenURL(gomock.Any(), url).
					Return(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, &err_repo.NotUniqueURLError{})
				s.EXPECT().FormatShortURL(expectedAlias).
					Return(testBaseURL + "/" + expectedAlias)
			},
		},
		{
			name:               "service failure (500)",
			url:                "https://example.com",
			expectedStatusCode: http.StatusInternalServerError,
			mockBehavior: func(s *mock_service.MockShortURL, url string) {
				// Имитируем любую системную ошибку
				s.EXPECT().ShortenURL(gomock.Any(), url).
					Return(link.ExpandedURL{}, errors.New("database connection failed"))

				// s.EXPECT().FormatShortURL НЕ вызовется, так как выполнение прервется на ошибке 500
			},
		},
	}
	// 🔸🔸Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)           // gomock сontroller
			svc := mock_service.NewMockShortURL(ctrl) // мок сервиса
			// 🔧настройка мока
			tt.mockBehavior(svc, tt.url)

			handler := Controller{ // хендлер с моком "внутри"
				shortener: svc,
			}
			r := chi.NewRouter()             // тестовый маршрутизатор
			r.Post("/", handler.ShortenText) // маршрут

			w := httptest.NewRecorder() // Имитация ResponseWriter
			// 🔧В httptest.NewRequest (Имитация запроса), target (второй аргумент) всегда "/..." (путь), он не может быть ""
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.url)) // тестируемый URL передаем в тело

			r.ServeHTTP(w, req) // Имитация работы HTTP‑сервера в памяти
			// 🔸🔸🔸Assert 🔧
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			// Если вернулся StatusCreated, то нет смысла проверять строку. Если StatusBadRequest тоже- нет строки ответа.
			// Отстается StatusConflict , но и тут не однозначна необходимость...
			if tt.expectedStatusCode == http.StatusConflict {
				// Сравниваем полученную строку с ожидаемой
				assert.Equal(t, testBaseURL+"/"+expectedAlias, w.Body.String())
				// // Другой вариант проверки полученной строки- можно проверить непустоту ответа или его длину.
				// assert.NotEmpty(t, w.Body.String())
			}
			// // Заголовки (Headers). Не понятно когда использовать..
			// assert.Equal(t, "text/plain", w.Header().Get("Content-Type"))
		})
	}
}

// 03.04.2026 Не ожиданно получил Тестирование всей цепочки (Integration-тест) для gzip
// Работает! Не знаю правильный ли он, оставлю
func TestController_GzipIntegration(t *testing.T) {
	ctrl := gomock.NewController(t)
	svc := mock_service.NewMockShortURL(ctrl)

	// Настраиваем роутер ТАК ЖЕ, как в основном приложении (с Middleware)
	handler := &Controller{shortener: svc}
	r := handler.InitRoutes(slog.Default()) // Используем твой метод InitRoutes!

	// 1. Готовим сжатые данные для запроса
	var buf bytes.Buffer
	gz := gzip.NewWriter(&buf)
	gz.Write([]byte("https://example.com"))
	gz.Close()

	// 2. Ожидаем вызовы (как обычно)
	svc.EXPECT().ShortenURL(gomock.Any(), "https://example.com").Return(link.ExpandedURL{ID: "abc123"}, nil)
	svc.EXPECT().FormatShortURL("abc123").Return("http://localhost:8080/abc123")

	// 3. Делаем запрос с нужными заголовками
	req := httptest.NewRequest(http.MethodPost, "/", &buf)
	req.Header.Set("Content-Encoding", "gzip") // Чтобы сработал DecompressRequest
	req.Header.Set("Accept-Encoding", "gzip")  // Чтобы сработал Compress

	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)

	// 4. Проверяем результат
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "gzip", w.Header().Get("Content-Encoding")) // Проверка, что сжатие ответа сработало
}
