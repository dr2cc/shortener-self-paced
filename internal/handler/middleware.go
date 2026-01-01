package handler

import (
	"compress/gzip"
	"io"
	"net/http"
)

// Если тело запроса сжато с помощью gzip, возвращает gzip reader,
// в противном случае возвращает request body (default reader)
func getDecompressedReader(r *http.Request) (io.Reader, error) {
	if r.Header.Get("Content-Encoding") == "gzip" {
		return gzip.NewReader(r.Body)
	}
	return r.Body, nil
}
