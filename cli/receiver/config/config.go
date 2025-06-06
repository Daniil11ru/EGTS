package config

/*
Описание конфигурационного файла
*/

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Host                string                       `yaml:"host"`
	Port                string                       `yaml:"port"`
	ConnTTL             int                          `yaml:"conn_ttl"`
	LogLevel            string                       `yaml:"log_level"`
	LogFilePath         string                       `yaml:"log_file_path"`
	LogMaxAgeDays       int                          `yaml:"log_max_age_days"`
	Store               map[string]map[string]string `yaml:"storage"`
	DBSaveMonthStart    int                          `yaml:"db_save_month_start"`
	DBSaveMonthEnd      int                          `yaml:"db_save_month_end"`
	PacketDataFieldName string                       `yaml:"packet_data_field_name"`
}

func (s *Settings) GetEmptyConnTTL() time.Duration {
	return time.Duration(s.ConnTTL) * time.Second
}
func (s *Settings) GetListenAddress() string {
	return s.Host + ":" + s.Port
}

func (s *Settings) GetLogLevel() log.Level {
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

func New(confPath string) (Settings, error) {
	c := Settings{}
	data, err := os.ReadFile(confPath)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}

	if c.PacketDataFieldName == "" {
		c.PacketDataFieldName = "packet_data"
	}

	if c.DBSaveMonthStart == 0 {
		c.DBSaveMonthStart = 5 // Default to May
	}
	if c.DBSaveMonthEnd == 0 {
		c.DBSaveMonthEnd = 9 // Default to September
	}

	if c.DBSaveMonthStart < 1 || c.DBSaveMonthStart > 12 || c.DBSaveMonthEnd < 1 || c.DBSaveMonthEnd > 12 {
		log.Errorf("Invalid DBSaveMonthStart (%d) or DBSaveMonthEnd (%d). Values must be between 1 and 12. Defaulting to May (5) and September (9).", c.DBSaveMonthStart, c.DBSaveMonthEnd)
		c.DBSaveMonthStart = 5
		c.DBSaveMonthEnd = 9
	}

	if c.DBSaveMonthStart > c.DBSaveMonthEnd {
		log.Errorf("DBSaveMonthStart (%d) cannot be after DBSaveMonthEnd (%d). Defaulting to May (5) and September (9).", c.DBSaveMonthStart, c.DBSaveMonthEnd)
		c.DBSaveMonthStart = 5
		c.DBSaveMonthEnd = 9
	}

	return c, err
}

func (s *Settings) GetDBSaveMonthStart() int {
	return s.DBSaveMonthStart
}

func (s *Settings) GetDBSaveMonthEnd() int {
	return s.DBSaveMonthEnd
}
