package main

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/parser"
	"GoNews/internal/server"
	"GoNews/internal/stopsignal"
	"GoNews/internal/storage/mongodb"
	"log/slog"
	"os"
)

func main() {

	// Инициализируем конфиг файл и логгер.
	logger.SetupLogger()
	cfg := config.MustLoad()
	slog.Debug("config file and logger initialized")

	// Инициализируем базу данных.
	st := mongodb.New(cfg)
	slog.Debug("storage initialized")
	defer st.Close()

	// Инициализируем и запускаем парсер RSS.
	parser := parser.New(cfg, st)
	slog.Debug("parser initialized")
	err := parser.Start()
	if err != nil {
		slog.Error("parser cannot start", logger.Err(err))
		st.Close()
		os.Exit(1)
	}

	// Инициализируем сервер, объявляем обработчики API и запускаем сервер.
	srv := server.New(cfg)
	srv.API(st)
	srv.Middleware()
	srv.Start()
	slog.Info("Server started")

	// Блокируем выполнение основной горутины и ожидаем сигнала прерывания.
	stopsignal.Stop()

	// После сигнала прерывания останавливаем парсер и сервер.
	go parser.Shutdown()
	srv.Shutdown()

	slog.Info("Server stopped")
}
