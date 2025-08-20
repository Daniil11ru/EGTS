package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/daniil11ru/egts/cli/receiver/api"
	apidomain "github.com/daniil11ru/egts/cli/receiver/api/domain"
	apirepo "github.com/daniil11ru/egts/cli/receiver/api/repository"
	"github.com/daniil11ru/egts/cli/receiver/config"
	"github.com/daniil11ru/egts/cli/receiver/server"
	"github.com/daniil11ru/egts/cli/receiver/server/domain"
	repo "github.com/daniil11ru/egts/cli/receiver/server/repository"
	"github.com/daniil11ru/egts/cli/receiver/source"
	"github.com/daniil11ru/egts/cli/receiver/util"
	"github.com/robfig/cron"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
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

	primarySource, err := source.NewDefaultPrimary(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", config.Store["host"], config.Store["user"], config.Store["password"], config.Store["database"], config.Store["port"], config.Store["sslmode"]))
	if err != nil {
		log.Fatalf("Не удалось инициализировать источник данных: %v", err)
		return
	}

	go runServer(primarySource, config)

	go runApi(primarySource, config.ApiPort)

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

func runServer(source source.Primary, config config.Config) {
	primaryRepository := repo.Primary{Source: source}

	savePacket := domain.SavePacket{
		PrimaryRepository:            primaryRepository,
		AddVehicleMovementMonthStart: config.GetSaveTelematicsDataMonthStart(),
		AddVehicleMovementMonthEnd:   config.GetSaveTelematicsDataMonthEnd(),
	}
	if err := savePacket.Initialize(); err != nil {
		log.Fatalf("Не удалось инициализировать кэш: %v", err)
		return
	}

	defer savePacket.Shutdown()

	optimizeGeometry := domain.OptimizeGeometry{PrimaryRepository: primaryRepository}
	c := cron.New()
	c.AddFunc(config.OptimizeGeometryCronExpression, func() { optimizeGeometry.Run() })
	c.Start()
	log.Info("Запланирована ежедневная оптимизация геометрии треков")

	for providerID, addr := range config.GetListenAddresses() {
		srv := server.NewServer(addr, config.GetEmptyConnectionTTL(), providerID, &savePacket)
		go func(a string, s *server.Server) {
			if err := s.Run(); err != nil {
				log.Fatalf("Не удалось запустить сервер на %s: %v", a, err)
			}
		}(addr, srv)
	}

	select {}
}

func runApi(source source.Primary, port int32) {
	businessDataRepository := apirepo.NewBusinessDataDefault(source)
	handler := api.NewHandler(businessDataRepository)
	additionalDataRepository := apirepo.NewAdditionalDataDefault(source)
	getApiKeys := apidomain.GetApiKeys{
		ApiKeysRepository: additionalDataRepository,
	}
	controller, err := api.NewController(handler, &getApiKeys)
	if err != nil {
		log.Fatalf("Не удалось создать контроллер API: %v", err)
		return
	}
	log.Infof("Запуск API на порту %d", port)
	err = controller.Run(port)
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
