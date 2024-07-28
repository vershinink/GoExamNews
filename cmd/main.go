package main

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/parser"
	"GoNews/internal/server"
	"GoNews/internal/storage/mongodb"
)

func main() {

	cfg := config.MustLoad()
	log := logger.MustLoad()
	log.Debug("config file and logger initialized")

	st := mongodb.New(cfg)
	log.Debug("storage initialized")
	defer st.Close()

	parser := parser.New(cfg, log, st)
	log.Debug("parser initialized")

	parser.Start()

	srv := server.New(cfg)
	srv.API(log, st)
	srv.Start()
	log.Info("Server started")

	select {}

}
