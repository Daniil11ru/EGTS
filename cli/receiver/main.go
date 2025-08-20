package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/daniil11ru/egts/cli/receiver/api"
	arepo "github.com/daniil11ru/egts/cli/receiver/api/repository"
	"github.com/daniil11ru/egts/cli/receiver/config"
	"github.com/daniil11ru/egts/cli/receiver/server"
	"github.com/daniil11ru/egts/cli/receiver/server/domain"
	srepo "github.com/daniil11ru/egts/cli/receiver/server/repository"
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

type ServerSettings struct {
	Host                           string
	Ports                          map[int32]int
	ConnectionTtl                  int
	SaveTelematicsDataMonthStart   int
	SaveTelematicsDataMonthEnd     int
	OptimizeGeometryCronExpression string
}

func (s *ServerSettings) GetEmptyConnectionTtl() time.Duration {
	return time.Duration(s.ConnectionTtl) * time.Second
}
func (s *ServerSettings) GetListenAddresses() map[int32]string {
	addresses := make(map[int32]string)
	for providerID, port := range s.Ports {
		addresses[providerID] = s.Host + ":" + strconv.Itoa(int(port))
	}
	return addresses
}

type ApiSettings struct {
	Port int
}

type LoggingSettings struct {
	LogLevel      string
	LogFilePath   string
	LogMaxAgeDays int
}

func (s *LoggingSettings) GetLogLevel() log.Level {
	var lvl log.Level

	switch s.LogLevel {
	case "DEBUG":
		lvl = log.DebugLevel
	case "INFO":
		lvl = log.InfoLevel
	case "WARN":
		lvl = log.WarnLevel
	case "ERROR":
		lvl = log.ErrorLevel
	default:
		lvl = log.InfoLevel
	}
	return lvl
}

func main() {
	configFilePath := ""
	flag.StringVar(&configFilePath, "c", "", "")
	flag.Parse()
	config, err := getConfig(configFilePath)
	if err != nil {
		log.Fatalf("Не удалось получить конфиг: %v", err)
		return
	}

	configureLogging(LoggingSettings{
		LogLevel:      config.LogLevel,
		LogFilePath:   config.LogFilePath,
		LogMaxAgeDays: config.LogMaxAgeDays,
	})

	err = applyMigrations(fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
		config.Store["user"], config.Store["password"], config.Store["host"], config.Store["port"], config.Store["database"], config.Store["sslmode"]), config.MigrationsPath)
	if err != nil {
		log.Fatalf("Не удалось применить миграции: %v", err)
		return
	}

	primarySource, err := source.NewDefaultPrimary(fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%s sslmode=%s", config.Store["host"], config.Store["user"], config.Store["password"], config.Store["database"], config.Store["port"], config.Store["sslmode"]))
	if err != nil {
		log.Fatalf("Не удалось инициализировать источник данных: %v", err)
		return
	}

	go runServer(primarySource, ServerSettings{
		Host:                           config.Host,
		Ports:                          config.Ports,
		ConnectionTtl:                  config.ConnectionTtl,
		SaveTelematicsDataMonthStart:   config.SaveTelematicsDataMonthStart,
		SaveTelematicsDataMonthEnd:     config.SaveTelematicsDataMonthEnd,
		OptimizeGeometryCronExpression: config.OptimizeGeometryCronExpression,
	})

	go runApi(primarySource, ApiSettings{
		Port: config.ApiPort,
	})

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

func configureLogging(settings LoggingSettings) {
	log.SetLevel(settings.GetLogLevel())

	consoleFmt := &log.TextFormatter{ForceColors: true, FullTimestamp: false}
	log.SetFormatter(consoleFmt)
	log.SetOutput(os.Stdout)

	if settings.LogFilePath != "" {
		logDir := filepath.Dir(settings.LogFilePath)
		if _, err := os.Stat(logDir); os.IsNotExist(err) {
			if err := os.MkdirAll(logDir, os.ModePerm); err != nil {
				log.Fatalf("Не получилось создать директорию для логов: %v", err)
			}
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   settings.LogFilePath,
			MaxSize:    100,
			MaxBackups: 366,
			MaxAge:     settings.LogMaxAgeDays,
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

func runServer(source source.Primary, settings ServerSettings) {
	primaryRepository := srepo.Primary{Source: source}

	savePacket, err := domain.NewSavePacket(
		primaryRepository,
		settings.SaveTelematicsDataMonthStart,
		settings.SaveTelematicsDataMonthEnd,
	)
	if err != nil {
		log.Fatalf("Не удалось инициализировать кэш: %v", err)
		return
	}

	defer savePacket.Shutdown()

	optimizeGeometry := domain.OptimizeGeometry{PrimaryRepository: primaryRepository}
	c := cron.New()
	c.AddFunc(settings.OptimizeGeometryCronExpression, func() { optimizeGeometry.Run() })
	c.Start()
	log.Info("Запланирована ежедневная оптимизация геометрии треков")

	for providerID, addr := range settings.GetListenAddresses() {
		srv := server.NewServer(addr, settings.GetEmptyConnectionTtl(), providerID, savePacket)
		go func(a string, s *server.Server) {
			if err := s.Run(); err != nil {
				log.Fatalf("Не удалось запустить сервер на %s: %v", a, err)
			}
		}(addr, srv)
	}

	select {}
}

func runApi(source source.Primary, apiSettings ApiSettings) {
	businessDataRepository := arepo.NewBusinessDataDefault(source)
	handler := api.NewHandler(businessDataRepository)
	additionalDataRepository := arepo.NewAdditionalDataDefault(source)
	controller, err := api.NewController(handler, additionalDataRepository)
	if err != nil {
		log.Fatalf("Не удалось создать контроллер API: %v", err)
		return
	}
	log.Infof("Запуск API на порту %d", apiSettings.Port)
	err = controller.Run(apiSettings.Port)
	if err != nil {
		log.Fatal(err)
	}
}

func applyMigrations(databaseUrl, migrationsPath string) error {
	m, err := migrate.New(migrationsPath, databaseUrl)
	if err != nil {
		return fmt.Errorf("ошибка инициализации миграций: %w", err)
	}
	defer m.Close()

	if err := m.Up(); err != nil {
		if errors.Is(err, migrate.ErrNoChange) {
			log.Info("Нет новых миграций для применения")
			return nil
		}
		var d migrate.ErrDirty
		if errors.As(err, &d) {
			return fmt.Errorf("миграции остановились на версии %d (dirty)", d.Version)
		}
		return fmt.Errorf("ошибка применения миграций: %w", err)
	}

	log.Info("Миграции успешно применены")
	return nil
}
