package postgresql

/*
Настройки, которые могут (а некоторые – должны) быть в конфиге для подключения хранилища:

host = "localhost"
port = "5432"
user = "postgres"
password = "postgres"
database = "receiver"
table = "point"
sslmode = "disable"
*/

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Connector struct {
	connection *sql.DB
	config     map[string]string
}

func (c *Connector) Init(cfg map[string]string) error {
	var (
		err error
	)
	if cfg == nil {
		return fmt.Errorf("некорректная ссылка на конфигурацию")
	}
	c.config = cfg
	connStr := fmt.Sprintf("dbname=%s host=%s port=%s user=%s password=%s sslmode=%s",
		c.config["database"], c.config["host"], c.config["port"], c.config["user"], c.config["password"], c.config["sslmode"])
	if c.connection, err = sql.Open("postgres", connStr); err != nil {
		return fmt.Errorf("ошибка подключения к PostgreSQL: %v", err)
	}

	if err = c.connection.Ping(); err != nil {
		return fmt.Errorf("PostgreSQL недоступен: %v", err)
	}
	return err
}

func (c *Connector) Save(msg interface{ ToBytes() ([]byte, error) }) error {
	if msg == nil {
		return fmt.Errorf("некорректная ссылка на пакет")
	}

	innerPkg, err := msg.ToBytes()
	if err != nil {
		return fmt.Errorf("ошибка сериализации пакета: %v", err)
	}

	packetDataFieldName := c.config["packet_data_field"]
	if packetDataFieldName == "" {
		log.Warnf("Ключ 'packet_data_field' не найден в конфигурации хранилища. Используется значение по умолчанию 'data'.")
		packetDataFieldName = "data"
	} else {
		log.Infof("Используется поле '%s' для хранения телематических данных.", packetDataFieldName)
	}

	vehicleMovementTableName := c.config["vehicle_movement_table"]
	if vehicleMovementTableName == "" {
		log.Warnf("Ключ 'vehicle_movement_table' не найдет в конфигурации хранилища. Используется значение по умолчению 'vehicle_movement'.")
		vehicleMovementTableName = "vehicle_movement_table"
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s) VALUES ($1)", vehicleMovementTableName, packetDataFieldName)
	if _, err = c.connection.Exec(insertQuery, innerPkg); err != nil {
		return fmt.Errorf("не удалось вставить запись: %v", err)
	}
	return nil
}

func (c *Connector) Close() error {
	return c.connection.Close()
}
