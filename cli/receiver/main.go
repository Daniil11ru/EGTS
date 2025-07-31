package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daniil11ru/egts/cli/receiver/config"
	connector "github.com/daniil11ru/egts/cli/receiver/connector/implementation"
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

	cron "github.com/robfig/cron/v3"
)

func main() {
	cfgFilePath := ""
	flag.StringVar(&cfgFilePath, "c", "", "")
	flag.Parse()

	if cfgFilePath == "" {
		log.Fatalf("Не задан путь до конфига")
	}

	config, err := config.New(cfgFilePath)
	if err != nil {
		log.Fatalf("Ошибка парсинга конфига: %v", err)
	}

	log.SetLevel(config.GetLogLevel())

	consoleFmt := &log.TextFormatter{ForceColors: true, FullTimestamp: false}
	log.SetFormatter(consoleFmt)
	log.SetOutput(os.Stdout)

	if config.LogFilePath != "" {
		logDir := filepath.Dir(config.LogFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
				log.Fatalf("Не получилось создать директорию для логов: %v", err)
			}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   config.LogFilePath,
			MaxSize:    100,
			MaxBackups: 366,
			MaxAge:     config.LogMaxAgeDays,
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
		config.Store["user"], config.Store["password"], config.Store["host"], config.Store["port"], config.Store["database"], config.Store["sslmode"])
	applyMigrations(config.MigrationsPath, dbURL)

	connector := connector.Connector{}
	if err := connector.Connect(config.Store); err != nil {
		log.Fatalf("Не удалось подключиться к хранилищу: %v", err)
		return
	}

	primarySource := source.PrimarySource{}
	primarySource.Initialize(&connector)

	primaryRepository := repository.PrimaryRepository{Source: &primarySource}

	savePacket := domain.SavePacket{
		PrimaryRepository:            primaryRepository,
		AddVehicleMovementMonthStart: config.GetSaveTelematicsDataMonthStart(),
		AddVehicleMovementMonthEnd:   config.GetSaveTelematicsDataMonthEnd(),
	}
	if err := savePacket.Initialize(); err != nil {
		log.Fatalf("Не удалось инициализировать кэш: %v", err)
	}

	getIPWhiteList := domain.GetIPWhiteList{PrimaryRepository: primaryRepository}

	defer savePacket.Shutdown()
	defer connector.Close()

	srv := server.New(config.GetListenAddress(), config.GetEmptyConnectionTTL(), &savePacket, getIPWhiteList)
	srv.Run()

	optimizeGeometry := domain.OptimizeGeometry{PrimaryRepository: primaryRepository}
	c := cron.New()
	c.AddFunc(config.OptimizeGeometryCronExpression, func() { optimizeGeometry.Run() })
	c.Start()
	select {}
}

func applyMigrations(databaseURL string, migrationsPath string) error {
	m, err := migrate.New(
		migrationsPath,
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
