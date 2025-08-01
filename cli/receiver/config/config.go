package config

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Config struct {
	Host                           string            `yaml:"host"`
	Port                           string            `yaml:"port"`
	ConnectionTTL                  int               `yaml:"connection_ttl"`
	LogLevel                       string            `yaml:"log_level"`
	LogFilePath                    string            `yaml:"log_file_path"`
	LogMaxAgeDays                  int               `yaml:"log_max_age_days"`
	Store                          map[string]string `yaml:"storage"`
	SaveTelematicsDataMonthStart   int               `yaml:"save_telematics_data_month_start"`
	SaveTelematicsDataMonthEnd     int               `yaml:"save_telematics_data_month_end"`
	OptimizeGeometryCronExpression string            `yaml:"optimize_geometry_cron_expression"`
	MigrationsPath                 string            `yaml:"migrations_path"`
}

func (s *Config) GetEmptyConnectionTTL() time.Duration {
	return time.Duration(s.ConnectionTTL) * time.Second
}
func (s *Config) GetListenAddress() string {
	return s.Host + ":" + s.Port
}

func (s *Config) GetLogLevel() log.Level {
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

func NewConfig(configPath string) (Config, error) {
	c := Config{}
	data, err := os.ReadFile(configPath)
	if err != nil {
		return c, err
	}
	err = yaml.Unmarshal(data, &c)
	if err != nil {
		return c, err
	}

	if c.SaveTelematicsDataMonthStart == 0 {
		c.SaveTelematicsDataMonthStart = 5 // Май
	}
	if c.SaveTelematicsDataMonthEnd == 0 {
		c.SaveTelematicsDataMonthEnd = 9 // Сентябрь
	}

	if c.SaveTelematicsDataMonthStart < 1 || c.SaveTelematicsDataMonthStart > 12 || c.SaveTelematicsDataMonthEnd < 1 || c.SaveTelematicsDataMonthEnd > 12 {
		log.Errorf("Некорректное значение SaveTelematicsDataMonthStart (%d) или SaveTelematicsDataMonthEnd (%d). Значение не должно быть меньше 1 и превышать 12. В качестве значений по умолчению взяты май (5) и сентябрь (9).", c.SaveTelematicsDataMonthStart, c.SaveTelematicsDataMonthEnd)
		c.SaveTelematicsDataMonthStart = 5
		c.SaveTelematicsDataMonthEnd = 9
	}

	if c.SaveTelematicsDataMonthStart > c.SaveTelematicsDataMonthEnd {
		log.Errorf("SaveTelematicsDataMonthStart (%d) не может превышать SaveTelematicsDataMonthEnd (%d). В качестве значений по умолчению взяты май (5) и сентябрь (9).", c.SaveTelematicsDataMonthStart, c.SaveTelematicsDataMonthEnd)
		c.SaveTelematicsDataMonthStart = 5
		c.SaveTelematicsDataMonthEnd = 9
	}

	return c, err
}

func (s *Config) GetSaveTelematicsDataMonthStart() int {
	return s.SaveTelematicsDataMonthStart
}

func (s *Config) GetSaveTelematicsDataMonthEnd() int {
	return s.SaveTelematicsDataMonthEnd
}
