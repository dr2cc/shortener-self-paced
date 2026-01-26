package handler

import "net/http"

// Ping is a health check endpoint.
func (c *Controller) Ping(w http.ResponseWriter, r *http.Request) {
	err := c.service.CheckHealth(r.Context())
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}
