package source

import (
	"errors"
	"time"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	pgconnector "github.com/daniil11ru/egts/cli/receiver/connector/postgresql"
	pgsaver "github.com/daniil11ru/egts/cli/receiver/repository/movement/source/postgresql"
	log "github.com/sirupsen/logrus"
)

var now = time.Now

var ErrInvalidStorage = errors.New("хранилище не найдено")
var ErrUnknownStorage = errors.New("хранилище не поддерживается")

type Store interface {
	connector.Connector
	Saver
}

type Saver interface {
	Save(message interface{ ToBytes() ([]byte, error) }, vehicleID int) error
}

type Repository struct {
	storages         []Saver
	DBSaveMonthStart int
	DBSaveMonthEnd   int
}

func (r *Repository) AddStore(s Saver) {
	r.storages = append(r.storages, s)
}

func (r *Repository) Save(m interface{ ToBytes() ([]byte, error) }, vehicleID int) error {
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
		log.Infof("Данные не сохранены – текущий месяц №%s вне заданного диапазона [%s;%s]", currentMonth.String(), startMonth.String(), endMonth.String())
		return nil
	}

	for _, store := range r.storages {
		if err := store.Save(m, vehicleID); err != nil {
			return err
		}
	}
	return nil
}

func (r *Repository) LoadStorages(storages map[string]map[string]string) error {
	if len(storages) == 0 {
		return ErrInvalidStorage
	}

	var connector connector.Connector
	var saver Saver
	for store, params := range storages {
		switch store {
		case "postgresql":
			connector = &pgconnector.Connector{}
			connector.Connect(storages["postgresql"])
			saver, _ = pgsaver.NewSaver(connector, params)
		default:
			return ErrUnknownStorage
		}

		if err := connector.Connect(params); err != nil {
			return err
		}

		r.AddStore(saver)
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
