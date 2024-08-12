// Пакет для работы с сервером и обработчиками API.
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
func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/", http.FileServerFS(webapp.Serve()))
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	}
}

// Posts возвращает слайс последних по дате постов в формате JSON.
func Posts(st storage.Interface) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Posts"

		slog.Info("new request to receive last posts")

		n, err := strconv.Atoi(r.PathValue("n"))
		if err != nil {
			slog.Error("incorrect posts number", slog.String("op", operation))
			http.Error(w, "incorrect posts number", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		posts, err := st.Posts(ctx, n)
		if err != nil {
			slog.Error("failed to receive posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		resp := respConv(posts)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		err = json.NewEncoder(w).Encode(resp)
		if err != nil {
			slog.Error("failed to encode posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
	}
}

// respConv преобразует получаемые из БД посты в структуры
// для клиентского приложения.
func respConv(posts []storage.Post) []Response {
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
	return resp
}
