package handler

import "net/http"

// ❌ 27.01.2026 Проверить и удалить
// // Ping is a health check endpoint.
// func (c *Controller) Ping(w http.ResponseWriter, r *http.Request) {
// 	err := c.service.CheckHealth(r.Context())
// 	if err != nil {
// 		http.Error(w, err.Error(), http.StatusInternalServerError)
// 	}
// }

// ♊ 27.01.2026
func (c *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	if err := c.service.CheckHealth(r.Context()); err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}
