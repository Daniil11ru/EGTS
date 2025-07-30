package implementation

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Settings struct {
	Driver   string
	Host     string
	Port     string
	User     string
	Password string
	Database string
	SSLMode  string
}

type Connector struct {
	connection *sql.DB
	settings   Settings
}

func getOptionValue(optionName string, optionDefaultValue string, settings map[string]string) string {
	optionValue := settings[optionName]
	if optionValue == "" {
		log.Warnf("Ключ '%s' не найден в конфигурации хранилища. Используется значение по умолчанию '%s'.", optionName, optionDefaultValue)
		optionValue = optionDefaultValue
	}

	return optionValue
}

func (c *Connector) FillSettings(settings map[string]string) {
	c.settings.Driver = getOptionValue("driver", "postgres", settings)
	c.settings.Host = getOptionValue("host", "localhost", settings)
	c.settings.Port = getOptionValue("port", "5432", settings)
	c.settings.User = getOptionValue("user", "postgres", settings)
	c.settings.Password = getOptionValue("password", "123", settings)
	c.settings.Database = getOptionValue("database", "egts", settings)
	c.settings.SSLMode = getOptionValue("sslmode", "disable", settings)
}

func (c *Connector) Connect(settings map[string]string) error {
	var err error
	if settings == nil {
		return fmt.Errorf("некорректная ссылка на конфигурацию")
	}

	c.FillSettings(settings)

	connStr := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=%s",
		c.settings.Database, c.settings.Host, c.settings.Port, c.settings.User, c.settings.Password, c.settings.SSLMode)

	if c.settings.Driver == "postgres" {
		if c.connection, err = sql.Open("postgres", connStr); err != nil {
			return fmt.Errorf("ошибка подключения к PostgreSQL: %v", err)
		}
	} else {
		return fmt.Errorf("неизвестный драйвер базы данных: %s", c.settings.Driver)
	}

	if err = c.connection.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL недоступен: %v", err)
	}
	return err
}

func (c *Connector) GetConnection() *sql.DB {
	return c.connection
}

func (c *Connector) Close() error {
	return c.connection.Close()
}
