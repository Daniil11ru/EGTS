package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daniil11ru/egts/cli/receiver/api"
	"github.com/daniil11ru/egts/cli/receiver/api/repository/implementation"
	"github.com/daniil11ru/egts/cli/receiver/api/source/postgre"
	"github.com/daniil11ru/egts/cli/receiver/config"
	connector "github.com/daniil11ru/egts/cli/receiver/connector/implementation"
	"github.com/daniil11ru/egts/cli/receiver/domain"
	repository "github.com/daniil11ru/egts/cli/receiver/repository/primary"
	"github.com/daniil11ru/egts/cli/receiver/server"
	source "github.com/daniil11ru/egts/cli/receiver/source/primary/pg"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/robfig/cron"
	"gorm.io/gorm"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"gorm.io/driver/postgres"
)

func main() {
	configFilePath := ""
	flag.StringVar(&configFilePath, "c", "", "")
	flag.Parse()
	config, err := getConfig(configFilePath)
	if err != nil {
		log.Fatalf("Не удалось получить конфиг: %v", err)
		return
	}

	configureLogging(config)

	applyMigrations(config)

	go runServer(config)

	go runApi(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", config.Host, config.Store["user"], config.Store["password"], config.Store["database"], config.Store["port"], config.Store["sslmode"]), 7001)

	select {}
}

func getConfig(configFilePath string) (config.Config, error) {
	var c config.Config
	var err error

	if configFilePath == "" {
		return c, &util.ErrorString{S: "не задан путь до конфига"}
	}

	c, err = config.NewConfig(configFilePath)
	if err != nil {
		return c, fmt.Errorf("ошибка парсинга конфига: %v", err)
	}

	return c, nil
}

func configureLogging(config config.Config) {
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
}

func runServer(config config.Config) {
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
		return
	}

	getIpWhiteList := domain.GetIPWhiteList{PrimaryRepository: primaryRepository}

	defer savePacket.Shutdown()
	defer connector.Close()

	optimizeGeometry := domain.OptimizeGeometry{PrimaryRepository: primaryRepository}
	c := cron.New()
	c.AddFunc(config.OptimizeGeometryCronExpression, func() { optimizeGeometry.Run() })
	c.Start()
	log.Info("Запланирована ежедневная оптимизация геометрии треков")

	server := server.NewServer(config.GetListenAddress(), config.GetEmptyConnectionTTL(), &savePacket, getIpWhiteList)
	err := server.Run()
	if err != nil {
		log.Fatalf("Не удалось запустить сервер: %v", err)
		return
	}
}

func runApi(dsn string, port int16) {
	db, _ := gorm.Open(postgres.New(postgres.Config{
		DSN:                  dsn,
		PreferSimpleProtocol: true,
	}), &gorm.Config{})
	source := postgre.New(db)
	repository := implementation.New(source)
	handler := api.NewHandler(repository)
	controller := api.NewController(handler)
	err := controller.Run(port)
	if err != nil {
		log.Fatal(err)
	}
}

func applyMigrations(config config.Config) error {
	databaseUrl := fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Store["user"], config.Store["password"], config.Store["host"], config.Store["port"], config.Store["database"], config.Store["sslmode"])

	m, err := migrate.New(
		config.MigrationsPath,
		databaseUrl,
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
