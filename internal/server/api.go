package server

import (
	"GoNews/internal/logger"
	"GoNews/internal/storage"
	"GoNews/webapp"
	"encoding/json"
	"log/slog"
	"net/http"
	"strconv"
)

// Index возвращает клиентское приложение.
func Index(log *slog.Logger, st storage.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

		log.Info("new request the index page")

		fs := http.StripPrefix("/", http.FileServer(webapp.Serve()))
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	}
}

// Posts возвращает слайс последних по дате постов в формате JSON.
func Posts(log *slog.Logger, st storage.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Posts"

		log.Info("new request to receive last posts")

		n, err := strconv.Atoi(r.PathValue("n"))
		if err != nil {
			log.Error("incorrect posts number", slog.String("op", operation))
			http.Error(w, "incorrect posts number", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		posts, err := st.Posts(ctx, n)
		if err != nil {
			log.Error("failed to receive posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(posts)
	}
}
