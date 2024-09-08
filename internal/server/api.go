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

// RespWeb - структура ответа, которую ожидает клиентское приложение
// из пакета webapp. Не используется в финальной версии итогового
// проекта.
type RespWeb struct {
	ID      int    `json:"ID"`
	Title   string `json:"Title"`
	Content string `json:"Content"`
	PubTime int64  `json:"PubTime"`
	Link    string `json:"Link"`
}

// Response - основная структура ответа сервера.
type Response struct {
	Pagination Pagination `json:"pagination"`
	Posts      []storage.Post
}

// Pagination - структура пагинации. Включает в себя общее число страниц,
// текущую страницу и число постов на странице.
type Pagination struct {
	Pages       int `json:"pages"`
	Currect     int `json:"current"`
	CountOnPage int `json:"countOnPage"`
}

const countOnPage int = 15

// Index возвращает клиентское приложение.
func Index() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		fs := http.StripPrefix("/", http.FileServerFS(webapp.Serve()))
		w.Header().Set("Content-Type", "text/html")
		w.Header().Set("Access-Control-Allow-Origin", "*")
		fs.ServeHTTP(w, r)
	}
}

// PostsWebApp записывает слайс последних n постов в ResponseWriter
// в формате JSON. Используется только в работе приложения из пакета webapp.
func PostsWebApp(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Posts"

		slog.Info("new request to receive last posts")

		n, err := strconv.Atoi(r.PathValue("n"))
		if err != nil {
			slog.Error("incorrect posts number", slog.String("op", operation))
			http.Error(w, "incorrect posts number", http.StatusBadRequest)
			return
		}

		opt := &storage.Options{Count: n}
		text := r.URL.Query().Get("s")
		if text != "" {
			opt.SearchQuery = text
		}

		ctx := r.Context()
		posts, err := st.Posts(ctx, opt)
		if err != nil {
			slog.Error("failed to receive posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}

		resp := respConv(posts)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(resp)
		if err != nil {
			slog.Error("failed to encode posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to encode posts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

// Posts записывает в ResponseWriter ответ Response в формате JSON.
// Ответ включает в себя объект пагинации и слайс соответствующих
// запросу постов из БД,
func Posts(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.PostsByPage"

		slog.Info("new request to receive posts")

		page, err := strconv.Atoi(r.URL.Query().Get("page"))
		if err != nil || page < 1 {
			page = 1
		}
		text := r.URL.Query().Get("s")

		opt := &storage.Options{}
		if text != "" {
			opt.SearchQuery = text
		}

		// Получаем общее количество постов, удовлетворяющих запросу.
		ctx := r.Context()
		num, err := st.Count(ctx, opt)
		if err != nil {
			slog.Error("failed to count posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}
		if num == 0 {
			slog.Error("posts not found", logger.Err(err), slog.String("op", operation))
			http.Error(w, "posts not found", http.StatusInternalServerError)
			return
		}

		// Высчитываем и формируем объект пагинации.
		pgCount := int(num) / countOnPage
		if page > int(pgCount) {
			page = 1
		}
		pg := Pagination{Pages: pgCount, Currect: page, CountOnPage: countOnPage}

		opt.Count = countOnPage
		opt.Offset = countOnPage * (page - 1)

		// Получаем посты для соответствующей страницы.
		posts, err := st.Posts(ctx, opt)
		if err != nil {
			slog.Error("failed to receive posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}

		resp := Response{Pagination: pg, Posts: posts}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		err = enc.Encode(resp)
		if err != nil {
			slog.Error("failed to encode posts", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to encode posts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

// PostByID записывает в ResponseWriter один пост по переданному ID.
func PostByID(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.PostByID"

		slog.Info("new request to receive post by ID")

		id := r.PathValue("id")
		if id == "" {
			slog.Error("empty post id", slog.String("op", operation))
			http.Error(w, "incorrect post id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		post, err := st.PostById(ctx, id)
		if err != nil {
			slog.Error("failed to receive post by id", slog.String("op", operation), slog.String("id", id))
			http.Error(w, "post not found", http.StatusBadRequest)
			return
		}

		w.Header().Set("Access-Control-Allow-Origin", "*")
		err = json.NewEncoder(w).Encode(post)
		if err != nil {
			slog.Error("failed to encode post", logger.Err(err), slog.String("op", operation))
			http.Error(w, "failed to encode post", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
	}
}

// respConv преобразует получаемые из БД посты в структуры
// для клиентского приложения.
func respConv(posts []storage.Post) []RespWeb {
	resp := make([]RespWeb, 0, len(posts))
	for k, p := range posts {
		var re RespWeb
		re.ID = k
		re.Title = p.Title
		re.Content = p.Content
		re.PubTime = p.PubTime.Unix()
		re.Link = p.Link
		resp = append(resp, re)
	}
	return resp
}
