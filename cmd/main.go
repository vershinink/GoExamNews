package main

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/parser"
	"GoNews/internal/server"
	"GoNews/internal/stopsignal"
	"GoNews/internal/storage/mongodb"
)

func main() {

	// Инициализируем конфиг файл и логгер.
	cfg := config.MustLoad()
	log := logger.MustLoad()
	log.Debug("config file and logger initialized")

	// Инициализируем базу данных.
	st := mongodb.New(cfg)
	log.Debug("storage initialized")
	defer st.Close()

	// Инициализируем и запускаем парсер RSS.
	parser := parser.New(cfg, log, st)
	log.Debug("parser initialized")
	parser.Start()

	// Инициализируем сервер, объявляем обработчики API и запускаем сервер.
	srv := server.New(cfg)
	srv.API(log, st)
	srv.Start()
	log.Info("Server started")

	// Блокируем выполнение основной горутины и ожидаем сигнала прерывания.
	stopsignal.Stop()

	// После сигнала прерывания останавливаем парсер и сервер.
	go parser.Shutdown()
	srv.Shutdown()

	log.Info("Server stopped")
}
