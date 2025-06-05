package storage

import (
	"errors"
	"time"

	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/mysql"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/nats"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/postgresql"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/rabbitmq"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/redis"
	"github.com/kuznetsovin/egts-protocol/cli/receiver/storage/store/tarantool_queue"
	log "github.com/sirupsen/logrus"
)

var now = time.Now

var ErrInvalidStorage = errors.New("хранилище не найдено")
var ErrUnknownStorage = errors.New("хранилище не поддерживается")

type Store interface {
	Connector
	Saver
}

type Saver interface {
	Save(interface{ ToBytes() ([]byte, error) }) error
}

type Connector interface {
	Init(map[string]string) error

	Close() error
}

type Repository struct {
	storages         []Saver
	DBSaveMonthStart int
	DBSaveMonthEnd   int
}

func (r *Repository) AddStore(s Saver) {
	r.storages = append(r.storages, s)
}

func (r *Repository) Save(m interface{ ToBytes() ([]byte, error) }) error {
	currentMonth := now().Month()
	startMonth := time.Month(r.DBSaveMonthStart)
	endMonth := time.Month(r.DBSaveMonthEnd)

	saveAllowed := false
	if startMonth <= endMonth {
		if currentMonth >= startMonth && currentMonth <= endMonth {
			saveAllowed = true
		}
	} else {
		if currentMonth >= startMonth || currentMonth <= endMonth {
			saveAllowed = true
		}
	}

	if !saveAllowed {
		log.Infof("Данные не сохранены – текущий месяц №%s вне заданного диапазана [%s;%s]", currentMonth.String(), startMonth.String(), endMonth.String())
		return nil
	}

	for _, store := range r.storages {
		if err := store.Save(m); err != nil {
			return err
		}
	}
	return nil
}

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

func NewRepositoryWithDefaults() *Repository {
	return &Repository{
		DBSaveMonthStart: 5,
		DBSaveMonthEnd:   9,
	}
}

func NewRepository(dbSaveMonthStart int, dbSaveMonthEnd int) *Repository {
	return &Repository{
		DBSaveMonthStart: dbSaveMonthStart,
		DBSaveMonthEnd:   dbSaveMonthEnd,
	}
}
