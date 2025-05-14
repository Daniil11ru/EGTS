package tarantool_queue

/*
Плагин для работы с Tarantool queue.

Раздел настроек, которые должны отвечають в конфиге для подключения хранилища:

host = "localhost"
port = "5672"
user = "user"
password = "pass"
max_recons = 5
timeout = 1
reconnect = 1
queue = "points"
*/

import (
	"fmt"
	"strconv"
	"time"

	"github.com/tarantool/go-tarantool"
	"github.com/tarantool/go-tarantool/queue"
)

type Connector struct {
	connection *tarantool.Connection
	queue      queue.Queue
	config     map[string]string
}

func (c *Connector) Init(cfg map[string]string) error {
	if cfg == nil {
		return fmt.Errorf("некорректная ссылка на конфигурацию")
	}

	c.config = cfg
	conStr := fmt.Sprintf("%s:%s", c.config["host"], c.config["port"])

	maxRecons, err := strconv.Atoi(c.config["max_recons"])
	if err != nil {
		return fmt.Errorf("не удалось получить MaxReconnects: %v", err)
	}
	timeout, err := strconv.Atoi(c.config["timeout"])
	if err != nil {
		return fmt.Errorf("не удалось получить timeout: %v", err)
	}
	reconnect, err := strconv.Atoi(c.config["reconnect"])
	if err != nil {
		return fmt.Errorf("не удалось получить reconnect: %v", err)
	}
	opts := tarantool.Opts{
		Timeout:       time.Duration(timeout) * time.Second,
		Reconnect:     time.Duration(reconnect) * time.Second,
		MaxReconnects: uint(maxRecons),
		User:          c.config["user"],
		Pass:          c.config["password"],
	}

	c.connection, err = tarantool.Connect(conStr, opts)
	if err != nil {
		return fmt.Errorf("не удалось подключиться к Tarantool: %v", err)
	}
	c.queue = queue.New(c.connection, c.config["queue"])

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

	_, err = c.queue.Put(innerPkg)
	if err != nil {
		return fmt.Errorf("не удалось отправить сообщение: %v", err)
	}
	return nil
}

func (c *Connector) Close() error {
	return c.connection.Close()
}
