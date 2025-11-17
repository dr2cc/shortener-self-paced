package ping

import (
	"app/internal/repository/pg"
	"app/internal/usecase/logger/sl"
	"log/slog"
	"net/http"

	_ "github.com/lib/pq"
	// _ {import} это импорт для "побочного эффекта" (side effect)
	// Значит, что нужны не все (основные) "эффекты" пакета,
	// а только дополнительные, нужные другим пакетам
	// (тут пакету "database/sql" нужен импорт драйвера)
)

func HealthCheck(repo *pg.PostgresRepo, log *slog.Logger) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		err := repo.DB.Ping()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			log.Error("Error connecting to the database:", sl.Err(err))
			return
		}
	}
}
