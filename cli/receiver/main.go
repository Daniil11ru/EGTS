package main

import (
	"flag"
	"os"
	"path/filepath"

	"github.com/daniil11ru/egts/cli/receiver/config"
	pgconnector "github.com/daniil11ru/egts/cli/receiver/connector/postgresql"
	"github.com/daniil11ru/egts/cli/receiver/domain"
	aux "github.com/daniil11ru/egts/cli/receiver/repository/auxiliary"
	mov "github.com/daniil11ru/egts/cli/receiver/repository/movement"
	"github.com/daniil11ru/egts/cli/receiver/repository/movement/source"
	"github.com/daniil11ru/egts/cli/receiver/server"
	auxsrc "github.com/daniil11ru/egts/cli/receiver/source/auxiliary/postgresql"

	"github.com/rifflock/lfshook"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
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

	telematicsDataStorages := source.NewRepository(cfg.GetDBSaveMonthStart(), cfg.GetDBSaveMonthEnd())
	if err := telematicsDataStorages.LoadStorages(cfg.Store); err != nil {
		log.Errorf("Ошибка загрузки хранилища: %v", err)
	}
	vehicleMovementRepository := mov.NewVehicleMovementRepository(telematicsDataStorages, 1024, 0)

	connector := pgconnector.Connector{}
	connector.Connect(cfg.Store["postgresql"])

	auxInfoSource := auxsrc.PostgresAuxSource{}
	auxInfoSource.Initialize(&connector)
	auxInfoRepository := aux.AuxiliaryInformationRepository{Source: &auxInfoSource}

	savePacket := domain.SavePackage{VehicleMovementRepository: vehicleMovementRepository, AuxiliaryInformationRepository: auxInfoRepository}

	getIPWhiteList := domain.GetIPWhiteList{AuxiliaryInformationRepository: auxInfoRepository}

	defer connector.Close()
	defer vehicleMovementRepository.Close()

	srv := server.NewCustom(cfg.GetListenAddress(), cfg.GetEmptyConnTTL(), &savePacket, getIPWhiteList)
	srv.Run()
}
