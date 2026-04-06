package handler

import (
	"app/internal/domain/link"
	err_repo "app/internal/lib/repository"
	mock_service "app/internal/service/mocks"
	"bytes"
	"compress/gzip"
	"encoding/json"
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
		queryBody          string
		expectedStatusCode int
		mockBehavior       mockBehavior
	}{
		{
			name:               "successful url shortening",
			queryBody:          "https://example.com",
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
			name:               "empty url",
			queryBody:          "",
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
			queryBody:          "https://example.com",
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
			queryBody:          "https://example.com",
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
			// 🔧настройка мока сервиса ShortURL
			tt.mockBehavior(svc, tt.queryBody)

			handler := Controller{ // хендлер с моком "внутри"
				shortener: svc,
			}
			r := chi.NewRouter() // тестовый маршрутизатор
			// 🎯Цель теста - проверка работы хендлера handler.ShortenText, на маршруте "/"
			r.Post("/", handler.ShortenText)

			w := httptest.NewRecorder() // Имитация ResponseWriter
			// 🔧В httptest.NewRequest (Имитация запроса), target (второй аргумент) всегда "/..." (путь), он не может быть ""
			// третий аргумент- объект Reader, считывающий данные из строки (наш queryBody)
			req := httptest.NewRequest(http.MethodPost, "/", strings.NewReader(tt.queryBody))

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

// 03.04.2026 Неожиданно получил Тестирование всей цепочки (Integration-тест) для gzip
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

func TestController_ShortenAPI(t *testing.T) {
	// 🔸Arrange
	type mockBehavior func(s *mock_service.MockShortURL, url string, mockError error, expectedJSON string)
	type queryBody struct {
		URL string `json:"url"`
	}

	tests := []struct {
		name               string
		requestBody        queryBody
		expectedStatusCode int
		mockError          error
		expectedJSON       string
		mockBehavior       mockBehavior
	}{
		{
			name: "successful url shortening",
			requestBody: queryBody{
				URL: "https://example.com",
			},
			expectedStatusCode: http.StatusCreated,
			mockError:          nil,
			expectedJSON:       `{"result":"http://localhost:8080/abc123"}`,
			mockBehavior: func(s *mock_service.MockShortURL, url string, mockError error, expectedJSON string) {
				gomock.InOrder( // устанавливает порядок вызовов. Метод InOrder не обязателен.
					// 1️⃣ При обращении к объекту s мы будем ОЖИДАТЬ().
					s.EXPECT().
						// 2️⃣ что вызов метода (структуры MockShortURLMockRecorder) ShortenURL(gomock.Any(), url)
						ShortenURL(gomock.Any(), url).
						// 3️⃣ тестируемому коду Вернет(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, nil)
						// (имитацию ответа от Controller.shortener.ShortenURL(...) (link.ExpandedURL, error))
						Return(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, mockError),
					//Return(url),
					// 2. СРАЗУ ЖЕ ожидаем вызов форматирования (вернет полный URL)
					// Следующим мы ОЖИДАЕМ().FormatShortURL()
					s.EXPECT().FormatShortURL(expectedAlias).
						Return(testBaseURL+"/"+expectedAlias), //expectedJSON),
				)
			},
		},
		{
			name: "empty url",
			requestBody: queryBody{
				URL: "",
			},
			expectedStatusCode: http.StatusBadRequest,
			mockBehavior: func(s *mock_service.MockShortURL, url string, mockError error, expectedJSON string) {
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
			name: "conflict - url already exists",
			requestBody: queryBody{
				URL: "https://example.com",
			},
			expectedStatusCode: http.StatusConflict,
			mockError:          &err_repo.NotUniqueURLError{}, // наша специфическая ошибка
			expectedJSON:       `{"result":"http://localhost:8080/abc123"}`,
			mockBehavior: func(s *mock_service.MockShortURL, url string, mockError error, expectedJSON string) {
				s.EXPECT().ShortenURL(gomock.Any(), url).
					Return(link.ExpandedURL{OriginalURL: url, ID: expectedAlias}, mockError)
				s.EXPECT().FormatShortURL(expectedAlias).
					Return(testBaseURL + "/" + expectedAlias)
			},
		},
		{
			name: "service failure (500)",
			requestBody: queryBody{
				URL: "https://example.com",
			},
			expectedStatusCode: http.StatusInternalServerError,
			mockError:          errors.New("database connection failed"), // Имитируем любую системную ошибку
			mockBehavior: func(s *mock_service.MockShortURL, url string, mockError error, expectedJSON string) {
				s.EXPECT().ShortenURL(gomock.Any(), url).
					Return(link.ExpandedURL{}, mockError)

				// s.EXPECT().FormatShortURL НЕ вызовется, так как выполнение прервется на ошибке 500
			},
		},
	}
	// 🔸🔸Act
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctrl := gomock.NewController(t)           // gomock сontroller
			svc := mock_service.NewMockShortURL(ctrl) // мок сервиса ShortURL
			// 🔧настройка мока сервиса ShortURL
			// "удача"
			// tt.requestBody.URL: "https://example.com",
			// mockError:          nil,
			// expectedJSON:       `{"result":"http://localhost:8080/abc123"}`,
			tt.mockBehavior(svc, tt.requestBody.URL, tt.mockError, tt.expectedJSON)
			// 📌Мок (здесь- сервиса ShortURL) должен предоставлять тестируемому объекту (здесь- handler.ShortenAPI)
			// те данные, что вернула бы зависимость (к примеру сервис shortener) при реальной работе, в конкретной ситуации ("успех", "не правильное тело" и т.д.)
			handler := Controller{ // хендлер с моком "внутри"
				shortener: svc,
			}

			w := httptest.NewRecorder() // Имитация ResponseWriter
			// 🔧Превращаем структуру в JSON-байты для отправки
			body, _ := json.Marshal(tt.requestBody)
			req := httptest.NewRequest(http.MethodPost, "/api/shorten", bytes.NewBuffer(body)) // Имитация запроса
			// Для JSON эндпоинта обязательно добавляем заголовок
			req.Header.Set("Content-Type", "application/json")

			r := chi.NewRouter() // тестовый маршрутизатор
			// 🎯Цель теста - проверка работы хендлера handler.ShortenAPI, на маршруте "/api/shorten"
			r.Post("/api/shorten", handler.ShortenAPI)
			r.ServeHTTP(w, req) // Имитация работы HTTP‑сервера в памяти
			// 🔸🔸🔸Assert 🔧
			assert.Equal(t, tt.expectedStatusCode, w.Code)
			// Если вернулся StatusCreated, то нет смысла проверять строку. Если StatusBadRequest тоже- нет строки ответа.
			// Отстается StatusConflict , но и тут не однозначна необходимость...
			if tt.expectedStatusCode == http.StatusConflict {
				// Сравниваем полученную строку с ожидаемой
				assert.Equal(t, tt.expectedJSON, w.Body.String())
				// // Другой вариант проверки полученной строки- можно проверить непустоту ответа или его длину.
				// assert.NotEmpty(t, w.Body.String())
			}
		})
	}
}
