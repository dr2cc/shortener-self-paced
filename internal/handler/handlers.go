// Package http_handlers contains functions that handles http requests
package handlers

import (
	"app/internal/config"
	service "app/internal/service"
	"app/internal/usecase/crypto"
	mwLogger "app/pkg/middleware/logger"
	"compress/flate"
	"compress/gzip"
	"encoding/hex"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

const UserIDCookieName = "shortener-user-id"

type Handler struct {
	Mux *chi.Mux // маршрутизатор, который мы будем использовать для обработки запросов
	// handler имеет (видимо везде!) в качестве зависимости указатель на структуру сервисоВ (по этому множественное число!)
	services *service.Shortener   // сервис, который содержит бизнес-логику (generator, Random), хранилище, конфигурацию
	crypto   crypto.Cryptographer // интерфейс, который будет использовать для шифрования и дешифрования значений
	log      *slog.Logger         // логгер
}

// NewHandler создает новый экземпляр структуры Handler, инициализирует chi мультиплексор,
// и выбирает службу
// и crypto
func NewHandler(service *service.Shortener, log *slog.Logger, config *config.Config) *Handler {
	cryptographer := crypto.GCMAESCryptographer{Key: config.EncryptionKey, Random: service.Random}
	return &Handler{
		Mux:      chi.NewMux(),
		services: service,
		crypto:   &cryptographer,
		log:      log,
	}
}

// Конструктор NewRouter создает новый маршрутизатор, добавляет middleware, а затем добавляет маршруты
func NewRouter(service *service.Shortener, cfg *config.Config, log *slog.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, log, cfg)

	router.Get("/{id}", h.Redirect)
	router.Post("/", h.ShortenText)
	// // При простой аутентификации, можно использовать такую конструкцию:
	// router.Route("/", func(r chi.Router) {
	// 	r.Use(middleware.BasicAuth("url-shortener", map[string]string{
	// 		cfg.User: cfg.Password,
	// 	}))
	// 	r.Post("/", h.ShortText)
	// })

	router.Post("/api/shorten", h.ShortenAPI)
	// iter10
	// Добавьте в сервис хендлер GET /ping,
	// который при запросе проверяет соединение с базой данных.
	// При успешной проверке хендлер должен вернуть HTTP-статус 200 OK, при неуспешной — 500 Internal Server Error.
	//
	router.Get("/ping", h.Ping)
	// iter12
	// Добавьте новый хендлер POST /api/shorten/batch,
	// принимающий в теле запроса множество URL для сокращения в формате:
	router.Post("/api/shorten/batch", h.BatchShortenAPI)

	//************************************************************************************
	// iter14
	// 	Добавьте в сервис функциональность аутентификации пользователя.
	// Сервис должен:
	// ◽ Выдавать пользователю симметрично подписанную куку, содержащую уникальный идентификатор пользователя,
	// если такой куки не существует или она не проходит проверку подлинности.

	// ◽ Иметь хендлер GET /api/user/urls,
	// который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
	// [
	//     {
	//         "short_url": "http://...",
	//         "original_url": "http://..."
	//     },
	//     ...
	// ]
	router.Get("/api/user/urls", h.UserURLs)
	//

	return router
}

// Если тело запроса сжато с помощью gzip, возвращает gzip reader,
// в противном случае возвращает request body (default reader)
func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

// addEncryptedUserIDToCookie encrypts the userID and setting it as a cookie.
func (h *Handler) addEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
	encryptedUserID, err := h.crypto.Encrypt([]byte(userID))
	if err != nil {
		return err
	}

	encodedCookieValue := hex.EncodeToString(encryptedUserID)

	http.SetCookie(
		*w,
		&http.Cookie{
			Name:  UserIDCookieName,
			Value: encodedCookieValue,
		},
	)
	return nil
}

// getUserID gets the userID from the cookie.
func (h *Handler) getUserID(r *http.Request) string {
	encodedCookie, err := r.Cookie(UserIDCookieName)
	if err != nil {
		return h.services.GenerateNewUserID()
	}

	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
	if err != nil {
		return h.services.GenerateNewUserID()
	}

	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
	if err != nil {
		return h.services.GenerateNewUserID()
	}

	return string(decryptedUserID)
}

// Ping is a health check endpoint.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.services.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
