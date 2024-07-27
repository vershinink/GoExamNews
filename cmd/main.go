package main

import (
	"GoNews/internal/config"
	"GoNews/internal/logger"
	"GoNews/internal/parser"
	"GoNews/internal/storage/mongodb"
	"os"
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

	err := parser.Start()
	if err != nil {
		log.Error("there are no correct rss links in the config file")
		log.Info("Parser cannot start")
		os.Exit(1)
	}

	log.Info("Parser started successfully")

	select {}
}
