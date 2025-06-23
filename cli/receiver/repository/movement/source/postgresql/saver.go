package postgresql

import (
	"fmt"

	connector "github.com/daniil11ru/egts/cli/receiver/connector"
	_ "github.com/lib/pq"
	log "github.com/sirupsen/logrus"
)

type Settings struct {
	PacketDataField      string
	VehicleMovementTable string
}

type Saver struct {
	connector connector.Connector
	settings  Settings
}

func getOptionValue(optionName string, optionDefaultValue string, settings map[string]string) string {
	optionValue := settings[optionName]
	if optionValue == "" {
		log.Warnf("Ключ '%s' не найден в конфигурации хранилища. Используется значение по умолчанию '%s'.", optionName, optionDefaultValue)
		optionValue = optionDefaultValue
	}

	return optionValue
}

func GetSettings(rawSettings map[string]string) Settings {
	settings := Settings{}
	settings.PacketDataField = getOptionValue("packet_data_field", "data", rawSettings)
	settings.VehicleMovementTable = getOptionValue("vehicle_movement_table", "vehicle_movement", rawSettings)

	return settings
}

func NewSaver(connector connector.Connector, rawSettings map[string]string) (*Saver, error) {
	if err := connector.GetConnection().Ping(); err != nil {
		return nil, fmt.Errorf("PostgreSQL недоступен: %v", err)
	}

	settings := GetSettings(rawSettings)
	saver := &Saver{connector: connector, settings: settings}

	return saver, nil
}

func (c *Saver) Save(msg interface{ ToBytes() ([]byte, error) }, vehicleID int) error {
	if msg == nil {
		return fmt.Errorf("некорректная ссылка на пакет")
	}

	innerPkg, err := msg.ToBytes()
	if err != nil {
		return fmt.Errorf("ошибка сериализации пакета: %v", err)
	}

	packetDataField := c.settings.PacketDataField
	vehicleMovementTable := c.settings.VehicleMovementTable

	insertQuery := fmt.Sprintf("INSERT INTO %s (%s, vehicle_id) VALUES ($1, $2)", vehicleMovementTable, packetDataField)
	if _, err = c.connector.GetConnection().Exec(insertQuery, innerPkg, vehicleID); err != nil {
		return fmt.Errorf("не удалось вставить запись: %v", err)
	}
	return nil
}
