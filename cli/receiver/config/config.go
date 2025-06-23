package config

import (
	"os"
	"time"

	log "github.com/sirupsen/logrus"

	"gopkg.in/yaml.v2"
)

type Settings struct {
	Host                         string            `yaml:"host"`
	Port                         string            `yaml:"port"`
	ConnTTL                      int               `yaml:"conn_ttl"`
	LogLevel                     string            `yaml:"log_level"`
	LogFilePath                  string            `yaml:"log_file_path"`
	LogMaxAgeDays                int               `yaml:"log_max_age_days"`
	Store                        map[string]string `yaml:"storage"`
	SaveTelematicsDataMonthStart int               `yaml:"save_telematics_data_month_start"`
	SaveTelematicsDataMonthEnd   int               `yaml:"save_telematics_data_month_end"`
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

	if c.SaveTelematicsDataMonthStart == 0 {
		c.SaveTelematicsDataMonthStart = 5 // Май
	}
	if c.SaveTelematicsDataMonthEnd == 0 {
		c.SaveTelematicsDataMonthEnd = 9 // Сентябрь
	}

	if c.SaveTelematicsDataMonthStart < 1 || c.SaveTelematicsDataMonthStart > 12 || c.SaveTelematicsDataMonthEnd < 1 || c.SaveTelematicsDataMonthEnd > 12 {
		log.Errorf("Invalid SaveTelematicsDataMonthStart (%d) or SaveTelematicsDataMonthEnd (%d). Values must be between 1 and 12. Defaulting to May (5) and September (9).", c.SaveTelematicsDataMonthStart, c.SaveTelematicsDataMonthEnd)
		c.SaveTelematicsDataMonthStart = 5
		c.SaveTelematicsDataMonthEnd = 9
	}

	if c.SaveTelematicsDataMonthStart > c.SaveTelematicsDataMonthEnd {
		log.Errorf("SaveTelematicsDataMonthStart (%d) cannot be after SaveTelematicsDataMonthEnd (%d). Defaulting to May (5) and September (9).", c.SaveTelematicsDataMonthStart, c.SaveTelematicsDataMonthEnd)
		c.SaveTelematicsDataMonthStart = 5
		c.SaveTelematicsDataMonthEnd = 9
	}

	return c, err
}

func (s *Settings) GetSaveTelematicsDataMonthStart() int {
	return s.SaveTelematicsDataMonthStart
}

func (s *Settings) GetSaveTelematicsDataMonthEnd() int {
	return s.SaveTelematicsDataMonthEnd
}
