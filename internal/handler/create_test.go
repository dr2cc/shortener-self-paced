package handler_test

import (
	"app/internal/config"
	"app/internal/domain/link"
	"app/internal/handler"
	handlers "app/internal/handler"
	"app/internal/lib/httpio"
	err_repo "app/internal/lib/repository"
	"app/internal/service"
	"context"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandler_ShortenText(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		service *service.Service
		config  *config.Config
		// Named input parameters for target function.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewHandler(tt.service)
			h.ShortenText(tt.w, tt.r)
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	// 1. Общая настройка
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	svc := &mockService{} // Твой мок-объект
	ctrl := handler.NewHandler(svc)

	// 2. Описание всех сценариев
	tests := []struct {
		name         string
		inputBody    string
		mockError    error
		expectedCode int
		expectedJSON string
	}{
		{
			name:         "Success 201",
			inputBody:    `{"url": "https://google.com"}`,
			mockError:    nil,
			expectedCode: http.StatusCreated,
			expectedJSON: `{"result":"http://localhost:8080/shortid"}`,
		},
		{
			name:         "Empty URL 400",
			inputBody:    `{"url": ""}`,
			mockError:    nil,
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"error":"url required"}`,
		},
		{
			name:         "Conflict 409",
			inputBody:    `{"url": "https://already-exists.com"}`,
			mockError:    &err_repo.NotUniqueURLError{}, // Имитируем конфликт
			expectedCode: http.StatusConflict,
			expectedJSON: `{"result":"http://localhost:8080/shortid"}`,
		},
		{
			name:         "Invalid JSON 400",
			inputBody:    `{"url": "bad-json"`, // Ошибка формата
			mockError:    nil,
			expectedCode: http.StatusBadRequest,
			expectedJSON: `{"error":"invalid json"}`,
		},
	}

	// 3. Запуск цикла тестов
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Настраиваем поведение мока под конкретный случай
			svc.nextError = tt.mockError

			req := httptest.NewRequest(http.MethodPost, "/api/shorten", strings.NewReader(tt.inputBody))
			req.Header.Set("Content-Type", "application/json")

			// Прокидываем логгер в контекст (имитация Middleware)
			ctx := context.WithValue(req.Context(), httpio.LoggerKey, logger)
			req = req.WithContext(ctx)

			w := httptest.NewRecorder()

			// ВЫЗОВ
			ctrl.ShortenAPI(w, req)

			// ПРОВЕРКИ
			assert.Equal(t, tt.expectedCode, w.Code)
			if tt.expectedJSON != "" {
				assert.JSONEq(t, tt.expectedJSON, w.Body.String())
			}
		})
	}
}

type mockService struct {
	nextError error
}

// Реализуем ShortURL часть
func (m *mockService) FormatShortURL(id string) string {
	return "http://localhost:8080/" + id
}

func (m *mockService) ShortenURL(ctx context.Context, url string) (link.ExpandedURL, error) {
	return link.ExpandedURL{ID: "shortid"}, m.nextError
}

func (m *mockService) ShortenBatch(ctx context.Context, batch []service.BatchInput) ([]link.ExpandedURL, error) {
	return []link.ExpandedURL{{ID: "id1"}, {ID: "id2"}}, m.nextError
}

func (m *mockService) FindURL(ctx context.Context, id string) (link.ExpandedURL, error) {
	return link.ExpandedURL{OriginalURL: "http://original.com", ID: id}, m.nextError
}

// Реализуем Pinger часть
func (m *mockService) CheckHealth(ctx context.Context) error {
	return m.nextError
}
