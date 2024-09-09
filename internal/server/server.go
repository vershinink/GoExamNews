package server

import (
	"GoNews/internal/config"
	"GoNews/internal/middleware"
	"GoNews/internal/storage"
	"context"
	"errors"
	"log"
	"log/slog"
	"net/http"
	"time"
)

// Server - структура сервера.
type Server struct {
	srv *http.Server
	mux *http.ServeMux
}

// New - конструктор сервера.
func New(cfg *config.Config) *Server {
	m := http.NewServeMux()
	server := &Server{
		srv: &http.Server{
			Addr:         cfg.Address,
			Handler:      m,
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			IdleTimeout:  cfg.IdleTimeout,
		},
		mux: m,
	}
	return server
}

// Start запускает HTTP сервер в отдельной горутине.
func (s *Server) Start(st storage.DB) {
	go func() {
		if err := s.srv.ListenAndServe(); err != nil {
			if errors.Is(err, http.ErrServerClosed) {
				return
			}
			slog.Error("failed to start server")
		}
	}()
}

// Middleware инициализирует все обработчики middleware.
func (s *Server) Middleware() {
	wrappedMux := middleware.RequestID(middleware.Logger(s.mux))
	s.srv.Handler = wrappedMux
}

// API инициализирует все обработчики API.
func (s *Server) API(st storage.DB) {
	// s.mux.HandleFunc("GET /", Index())
	s.mux.HandleFunc("GET /news/id/{id}", PostByID(st))
	s.mux.HandleFunc("GET /news/{n}", PostsWebApp(st))
	s.mux.HandleFunc("GET /news", Posts(st))
}

// Shutdown останавливает сервер используя graceful shutdown.
func (s *Server) Shutdown() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := s.srv.Shutdown(ctx); err != nil {
		log.Fatalf("failed to stop server: %s", err.Error())
	}
}
