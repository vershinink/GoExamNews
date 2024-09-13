// Пакет для работы с сервером и обработчиками API.
package server

import (
	"GoNews/internal/logger"
	"GoNews/internal/middleware"
	"GoNews/internal/storage"
	"GoNews/webapp"
	"encoding/json"
	"errors"
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
	Pages   int `json:"pages"`
	Currect int `json:"current"`
	OnPage  int `json:"onPage"`
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
		const operation = "server.PostsWebApp"

		log := slog.Default().With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request to receive last posts")

		n, err := strconv.Atoi(r.PathValue("n"))
		if err != nil {
			log.Error("incorrect posts number")
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
			log.Error("failed to receive posts", logger.Err(err))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}
		log.Debug("posts received successfully", slog.Int("num", len(posts)))

		resp := respConv(posts)

		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(resp)
		if err != nil {
			log.Error("failed to encode posts", logger.Err(err))
			http.Error(w, "failed to encode posts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Info("request served successfuly")
		log = nil
	}
}

// Posts записывает в ResponseWriter ответ Response в формате JSON.
// Ответ включает в себя объект пагинации и слайс соответствующих
// запросу постов из БД,
func Posts(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.Posts"

		log := slog.Default().With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request to receive posts")

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
			log.Error("failed to count posts", logger.Err(err))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}
		if num == 0 {
			log.Error("posts not found")
			http.Error(w, "posts not found", http.StatusNotFound)
			return
		}
		log.Debug("posts count successfully", slog.Int64("num", num))

		// Высчитываем и формируем объект пагинации.
		pgCount := int(num) / countOnPage
		if int(num)%countOnPage != 0 {
			pgCount++
		}
		if page > int(pgCount) {
			page = 1
		}
		onPage := int(num) - (page-1)*countOnPage
		if onPage > 15 {
			onPage = 15
		}
		pg := Pagination{Pages: pgCount, Currect: page, OnPage: onPage}

		opt.Count = countOnPage
		opt.Offset = countOnPage * (page - 1)

		// Получаем посты для соответствующей страницы.
		posts, err := st.Posts(ctx, opt)
		if err != nil {
			log.Error("failed to receive posts", logger.Err(err))
			http.Error(w, "failed to receive posts from DB", http.StatusInternalServerError)
			return
		}
		log.Debug("posts received successfully", slog.Int("num", len(posts)))

		resp := Response{Pagination: pg, Posts: posts}
		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(resp)
		if err != nil {
			log.Error("failed to encode posts", logger.Err(err))
			http.Error(w, "failed to encode posts", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Info("request served successfuly")
		log = nil
	}
}

// PostByID записывает в ResponseWriter один пост по переданному ID.
func PostByID(st storage.DB) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		const operation = "server.PostByID"

		log := slog.Default().With(
			slog.String("op", operation),
			slog.String("request_id", middleware.GetReqID(r.Context())),
		)

		log.Info("request to receive post by ID")

		id := r.PathValue("id")
		if id == "" {
			log.Error("empty post id")
			http.Error(w, "incorrect post id", http.StatusBadRequest)
			return
		}

		ctx := r.Context()
		post, err := st.PostById(ctx, id)
		if err != nil {
			log.Error("failed to receive post by id", slog.String("id", id), logger.Err(err))
			if errors.Is(err, storage.ErrNotFound) {
				http.Error(w, "post not found", http.StatusNotFound)
				return
			}
			if errors.Is(err, storage.ErrIncorrectId) {
				http.Error(w, "incorrect post id", http.StatusBadRequest)
				return
			}
			http.Error(w, "failed to receive post", http.StatusInternalServerError)
			return
		}
		log.Debug("post by ID received successfully", slog.String("id", id))

		w.Header().Set("Access-Control-Allow-Origin", "*")
		enc := json.NewEncoder(w)
		enc.SetIndent("", "\t")
		err = enc.Encode(post)
		if err != nil {
			log.Error("failed to encode post", logger.Err(err))
			http.Error(w, "failed to encode post", http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		log.Info("request served successfuly", slog.String("id", id))
		log = nil
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
