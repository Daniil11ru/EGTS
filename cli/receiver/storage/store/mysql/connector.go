package mysql

/*
Плагин для работы с MySQL.
Плагин сохраняет пакет в jsonb поле point у заданной в настройках таблице.

Раздел настроек, которые должны отвечають в конфиге для подключения хранилища:

uri = "user:password@host:port/dbname"
database = "receiver"
table = "points"
*/

import (
	"database/sql"
	"fmt"

	_ "github.com/go-sql-driver/mysql"
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

	if c.connection, err = sql.Open("mysql", c.config["uri"]); err != nil {
		return fmt.Errorf("ошибка подключения к MySQL: %v", err)
	}

	if err = c.connection.Ping(); err != nil {
		return fmt.Errorf("MySQL недоступен: %v", err)
	}
	return err
}

func (c *Connector) Save(msg interface{ ToBytes() ([]byte, error) }) error {
	if msg == nil {
		return fmt.Errorf("некорректная ссылка на пакет")
	}

	innerPkg, err := msg.ToBytes()
	if err != nil {
		return fmt.Errorf("ошибка сериализации  пакета: %v", err)
	}

	insertQuery := fmt.Sprintf("INSERT INTO %s (point) VALUES (?)", c.config["table"])
	if _, err = c.connection.Exec(insertQuery, innerPkg); err != nil {
		return fmt.Errorf("не удалось вставить запись в MySQL: %v", err)
	}
	return nil
}

func (c *Connector) Close() error {
	return c.connection.Close()
}
