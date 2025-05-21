package main

import (
	"flag"
	"io"
	"os"
	"path/filepath"

	"github.com/kuznetsovin/egts-protocol/cli/receiver/config"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/server"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

func main() {
	cfgFilePath := ""
	flag.StringVar(&cfgFilePath, "c", "", "Конфигурационный файл")
	flag.Parse()

	if cfgFilePath == "" {
		log.Fatalf("Не задан путь до конфига")
	}

	cfg, err := config.New(cfgFilePath)
	if err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
	}

	log.SetLevel(cfg.GetLogLevel())

	// Configure file-based logging with rotation
	if cfg.LogFilePath != "" {
		logDir := filepath.Dir(cfg.LogFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
				log.Fatalf("Failed to create log directory: %v", err)
			}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.LogFilePath,
			MaxSize:    100, // megabytes
			MaxBackups: 3,
			MaxAge:     cfg.LogMaxAgeDays, //days
			Compress:   true,              // disabled by default
		}

		// Output to both stdout and file
		mw := io.MultiWriter(os.Stdout, lumberjackLogger)
		log.SetOutput(mw)
	}


	storages := storage.NewRepository()
	if err := storages.LoadStorages(cfg.Store); err != nil {
		log.Errorf("ошибка загрузка хранилища: %v", err)

		// TODO: clear after test
		store := storage.LogConnector{}
		if err := store.Init(nil); err != nil {
			log.Fatal(err)
		}

		storages.AddStore(store)
		defer store.Close()
	}

	srv := server.NewCustom(cfg.GetListenAddress(), cfg.GetEmptyConnTTL(), storages)

	srv.Run()
}
