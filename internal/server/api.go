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

// Response - структура ответа, которую ожидает клиентское приложение.
type Response struct {
	ID      int    `json:"ID"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
	PubTime int64  `json:"PubTime"`
	Link    string `json:"Link"`
}

// Index возвращает клиентское приложение.
func Index(log *slog.Logger, st storage.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {

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

		resp := make([]Response, 0, len(posts))
		for k, p := range posts {
			var re Response
			re.ID = k
			re.Title = p.Title
			re.Content = p.Content
			re.PubTime = p.PubTime.Unix()
			re.Link = p.Link
			resp = append(resp, re)
		}

		w.Header().Set("Content-Type", "application/json")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
	}
}
