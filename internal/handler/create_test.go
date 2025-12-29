package handlers_test

import (
	"app/internal/config"
	handlers "app/internal/handler"
	"app/internal/service"
	"log/slog"
	"net/http"
	"testing"
)

func TestHandler_ShortenText(t *testing.T) {
	// shorten_test

	// Arrange
	type want struct {
		statusCode int
		body       string
	}

	// Табличное тестирование (table-driven tests).
	// Создается переменная tests ,
	// ее тип - «срез из анонимных структур,
	// содержащих поля name, body, method, want
	tests := []struct {
		name   string
		want   want
		method string
		body   string
	}{
		{
			name: "post with url",
			want: want{
				statusCode: http.StatusCreated,
				body:       "http://localhost:8080/id",
			},
			method: http.MethodPost,
			body:   "url",
		},
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// h := handlers.NewHandler(tt.service, tt.log, tt.config)
			// h.ShortenText(tt.w, tt.r)
		})
	}
}

func TestHandler_ShortenAPI(t *testing.T) {
	tests := []struct {
		name string // description of this test case
		// Named input parameters for receiver constructor.
		// Именованные входные параметры для конструктора приемника.
		service *service.Shortener
		log     *slog.Logger
		config  *config.Config
		// Named input parameters for target function.
		// Именованные входные параметры для целевой функции.
		w http.ResponseWriter
		r *http.Request
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := handlers.NewHandler(tt.service, tt.log, tt.config)
			h.ShortenAPI(tt.w, tt.r)
		})
	}
}
