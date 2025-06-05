package storage

import (
	"errors"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/mysql"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/nats"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/postgresql"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/rabbitmq"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/redis"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/tarantool_queue"
)

var now = time.Now // For mocking time.Now() in tests

var ErrInvalidStorage = errors.New("storage not found")
var ErrUnknownStorage = errors.New("storage isn't support yet")

type Store interface {
	Connector
	Saver
}

// Saver интерфейс для подключения внешних хранилищ
type Saver interface {
	// Save сохранение в хранилище
	Save(interface{ ToBytes() ([]byte, error) }) error
}

// Connector интерфейс для подключения внешних хранилищ
type Connector interface {
	// Init установка соединения с хранилищем
	Init(map[string]string) error

	// Close закрытие соединения с хранилищем
	Close() error
}

// Repository набор выходных хранилищ
type Repository struct {
	storages         []Saver
	DBSaveMonthStart int
	DBSaveMonthEnd   int
}

// AddStore добавляет хранилище для сохранения данных
func (r *Repository) AddStore(s Saver) {
	r.storages = append(r.storages, s)
}

// Save сохраняет данные во все установленные хранилища
func (r *Repository) Save(m interface{ ToBytes() ([]byte, error) }) error {
	currentMonth := now().Month()
	startMonth := time.Month(r.DBSaveMonthStart)
	endMonth := time.Month(r.DBSaveMonthEnd)

	saveAllowed := false
	if startMonth <= endMonth {
		if currentMonth >= startMonth && currentMonth <= endMonth {
			saveAllowed = true
		}
	} else { // Wraps around year-end (e.g. November to February)
		if currentMonth >= startMonth || currentMonth <= endMonth {
			saveAllowed = true
		}
	}

	if !saveAllowed {
		log.Infof("Data not saved. Current month %s is outside the configured range [%s - %s]", currentMonth.String(), startMonth.String(), endMonth.String())
		return nil
	}

	for _, store := range r.storages {
		if err := store.Save(m); err != nil {
			return err
		}
	}
	return nil
}

// LoadStorages загружает хранилища из структуры конфига
func (r *Repository) LoadStorages(storages map[string]map[string]string) error {
	if len(storages) == 0 {
		return ErrInvalidStorage
	}

	var db Store
	for store, params := range storages {
		switch store {
		case "rabbitmq":
			db = &rabbitmq.Connector{}
		case "postgresql":
			db = &postgresql.Connector{}
		case "nats":
			db = &nats.Connector{}
		case "tarantool_queue":
			db = &tarantool_queue.Connector{}
		case "redis":
			db = &redis.Connector{}
		case "mysql":
			db = &mysql.Connector{}
		default:
			return ErrUnknownStorage
		}

		if err := db.Init(params); err != nil {
			return err
		}

		r.AddStore(db)
	}
	return nil
}

// NewRepository создает пустой репозиторий
func NewRepository(dbSaveMonthStart int, dbSaveMonthEnd int) *Repository {
	return &Repository{
		DBSaveMonthStart: dbSaveMonthStart,
		DBSaveMonthEnd:   dbSaveMonthEnd,
	}
}
