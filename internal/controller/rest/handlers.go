// Package http_handlers contains functions that handles http requests
package handlers

import (
	"app/internal/config"
	mwLogger "app/internal/usecase/middleware/logger"
	services "app/internal/usecase/shortener"
	"compress/flate"
	"compress/gzip"
	"io"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
)

type Handler struct {
	Mux     *chi.Mux            // router that we'll be using to handle our requests
	service *services.Shortener // service that will contain main business logic
	// crypto  crypto.Cryptographer // interface that we'll use to encrypt and decrypt values
}

// NewHandler creates a new instance of the Handler struct, initializes the chi mux, and sets the service and crypto fields
func NewHandler(service *services.Shortener, config *config.Config) *Handler {
	// cryptographer := crypto.GCMAESCryptographer{Key: config.EncryptionKey, Random: service.Random}
	return &Handler{
		Mux:     chi.NewMux(),
		service: service,
		// crypto:  &cryptographer,
	}
}

// // NewRouter creates a new router, adds some middleware, and then adds some routes
func NewRouter(service *services.Shortener, cfg *config.Config, log *slog.Logger) chi.Router {
	router := chi.NewRouter()

	router.Use(middleware.RequestID)
	router.Use(middleware.Logger)
	router.Use(mwLogger.New(log))
	router.Use(middleware.Recoverer)
	router.Use(middleware.Compress(flate.BestSpeed))

	h := NewHandler(service, cfg)

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
	// // iter10
	// // Добавьте в сервис хендлер GET /ping,
	// // который при запросе проверяет соединение с базой данных.
	// // При успешной проверке хендлер должен вернуть HTTP-статус 200 OK, при неуспешной — 500 Internal Server Error.
	// //
	// r.Get("/ping", h.Ping)
	// // iter12
	// // Добавьте новый хендлер POST /api/shorten/batch,
	// // принимающий в теле запроса множество URL для сокращения в формате:
	// r.Post("/api/shorten/batch", h.ShortenBatchAPI)

	// //************************************************************************************
	// // iter14
	// // 	Добавьте в сервис функциональность аутентификации пользователя.

	// // Сервис должен иметь хендлер GET /api/user/urls,
	// // который сможет вернуть пользователю все когда-либо сокращённые им URL в формате:
	// // [
	// //     {
	// //         "short_url": "http://...",
	// //         "original_url": "http://..."
	// //     },
	// //     ...
	// // ]
	// r.Get("/api/user/urls", h.UserURLs)
	// //

	// //
	// // здешний iter14
	// // Далее добавьте в сервис новый асинхронный хендлер DELETE /api/user/urls,
	// // который принимает список идентификаторов сокращённых URL для удаления в формате:
	// r.Delete("/api/user/urls", h.DeleteUrls)
	// //
	// r.Group(func(r chi.Router) {
	// 	r.Use(FromTrustedSubnet(ipChecker))
	// 	r.Get("/api/internal/stats", h.Stats)
	// })

	return router
}

// func FromTrustedSubnet(checkerInterface services.IPCheckerInterface) func(http.Handler) http.Handler {
// 	return func(next http.Handler) http.Handler {
// 		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
// 			fromTrustedSubnet, err := checkerInterface.IsRequestFromTrustedSubnet(r)
// 			if err != nil {
// 				http.Error(w, err.Error(), http.StatusForbidden)
// 				return
// 			}

// 			if !fromTrustedSubnet {
// 				http.Error(w, "forbidden", http.StatusForbidden)
// 				return
// 			}

// 			next.ServeHTTP(w, r)
// 		})
// 	}
// }

// If the request body is gzipped, return a gzip reader, otherwise return the request body (default reader)
func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}

// Ping is a health check endpoint.
func (h *Handler) Ping(w http.ResponseWriter, r *http.Request) {
	err := h.service.HealthCheck(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// // addEncryptedUserIDToCookie encrypts the userID and setting it as a cookie.
// func (h *Handler) addEncryptedUserIDToCookie(w *http.ResponseWriter, userID string) error {
// 	encryptedUserID, err := h.crypto.Encrypt([]byte(userID))
// 	if err != nil {
// 		return err
// 	}

// 	encodedCookieValue := hex.EncodeToString(encryptedUserID)

// 	http.SetCookie(
// 		*w,
// 		&http.Cookie{
// 			Name:  UserIDCookieName,
// 			Value: encodedCookieValue,
// 		},
// 	)
// 	return nil
// }

// // getUserID gets the userID from the cookie.
// func (h *Handler) getUserID(r *http.Request) string {
// 	encodedCookie, err := r.Cookie(UserIDCookieName)
// 	if err != nil {
// 		return h.service.GenerateNewUserID()
// 	}

// 	decodedCookie, err := hex.DecodeString(encodedCookie.Value)
// 	if err != nil {
// 		return h.service.GenerateNewUserID()
// 	}

// 	decryptedUserID, err := h.crypto.Decrypt(decodedCookie)
// 	if err != nil {
// 		return h.service.GenerateNewUserID()
// 	}

// 	return string(decryptedUserID)
// }
