package handlers

import (
	"app/internal/usecase/random"
	"log/slog"
	"net/http"
)

// save.New (expand)
func (h Handler) New(randomKey random.RandomGenerator, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// url, err := io.ReadAll(r.Body)
		// if err != nil {
		// 	http.Error(w, err.Error(), http.StatusInternalServerError)
		// 	return
		// }

		// if string(url) == "" {
		// 	http.Error(w, "content required", http.StatusBadRequest)
		// 	return
		// }
		// defer r.Body.Close()

		// shortUrl := entity.ShortURL{
		// 	OriginalURL: string(url),
		// 	ID:          randomKey.NewRandomString(),
		// }

		// // TODO: Шаг 2- pg
		// err = pg.CreateRecord(log, shortUrl, h.Repo)
		// if err != nil {
		// 	log.Error(err.Error())
		// 	http.Error(w, "failed to add record", http.StatusBadRequest)
		// 	return
		// }
	}
}
