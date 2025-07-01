package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daniil11ru/egts/cli/receiver/config"
	connector "github.com/daniil11ru/egts/cli/receiver/connector/postgresql"
	"github.com/daniil11ru/egts/cli/receiver/domain"
	repository "github.com/daniil11ru/egts/cli/receiver/repository/primary"
	"github.com/daniil11ru/egts/cli/receiver/server"
	source "github.com/daniil11ru/egts/cli/receiver/source/primary/pg"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
)

func main() {
	cfgFilePath := ""
	flag.StringVar(&cfgFilePath, "c", "", "")
	flag.Parse()

	if cfgFilePath == "" {
		log.Fatalf("Не задан путь до конфига")
	}

	cfg, err := config.New(cfgFilePath)
	if err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
	}

	log.SetLevel(cfg.GetLogLevel())

	consoleFmt := &log.TextFormatter{ForceColors: true, FullTimestamp: false}
	log.SetFormatter(consoleFmt)
	log.SetOutput(os.Stdout)

	if cfg.LogFilePath != "" {
		logDir := filepath.Dir(cfg.LogFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
				log.Fatalf("Не получилось создать директорию для логов: %v", err)
			}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.LogFilePath,
			MaxSize:    100,
			MaxBackups: 366,
			MaxAge:     cfg.LogMaxAgeDays,
			Compress:   true,
		}

		fileFmt := &log.TextFormatter{DisableColors: true, FullTimestamp: true}
		hook := lfshook.NewHook(lfshook.WriterMap{
			log.PanicLevel: lumberjackLogger,
			log.FatalLevel: lumberjackLogger,
			log.ErrorLevel: lumberjackLogger,
			log.WarnLevel:  lumberjackLogger,
			log.InfoLevel:  lumberjackLogger,
			log.DebugLevel: lumberjackLogger,
			log.TraceLevel: lumberjackLogger,
		}, fileFmt)

		log.AddHook(hook)
	}

	dbURL := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		cfg.Store["user"], cfg.Store["password"], cfg.Store["host"], cfg.Store["port"], cfg.Store["database"], cfg.Store["sslmode"])
	applyMigrations(dbURL)

	connector := connector.Connector{}
	connector.Connect(cfg.Store)

	primarySource := source.PrimarySource{}
	primarySource.Initialize(&connector)

	primaryRepository := repository.PrimaryRepository{Source: &primarySource}

	savePacket := domain.SavePacket{
		PrimaryRepository:            primaryRepository,
		AddVehicleMovementMonthStart: cfg.GetSaveTelematicsDataMonthStart(),
		AddVehicleMovementMonthEnd:   cfg.GetSaveTelematicsDataMonthEnd(),
	}
	savePacket.Initialize()
	getIPWhiteList := domain.GetIPWhiteList{PrimaryRepository: primaryRepository}

	defer connector.Close()

	srv := server.New(cfg.GetListenAddress(), cfg.GetEmptyConnectionTTL(), &savePacket, getIPWhiteList)
	srv.Run()
}

func applyMigrations(databaseURL string) error {
	m, err := migrate.New(
		"file:///app/migrations",
		databaseURL,
	)
	if err != nil {
		return fmt.Errorf("ошибка инициализации миграций: %v", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if err == migrate.ErrNoChange {
			log.Info("Нет новых миграций для применения")
			return nil
		}
		return fmt.Errorf("ошибка применения миграций: %v", err)
	}

	log.Info("Миграции успешно применены")
	return nil
}
