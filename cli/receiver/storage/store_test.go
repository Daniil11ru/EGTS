package storage

import (
	"io"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

type mockSaver struct {
	saveCalled bool
}

func (ms *mockSaver) Save(data interface{ ToBytes() ([]byte, error) }) error {
	ms.saveCalled = true
	return nil
}

type testData struct{}

func (td testData) ToBytes() ([]byte, error) {
	return []byte("test"), nil
}

func TestRepository_Save_DateLogic(t *testing.T) {
	log.SetOutput(io.Discard)

	// Тестовые данные
	dummyData := testData{}

	tests := []struct {
		name              string
		repoStartMonth    int
		repoEndMonth      int
		mockedCurrentTime time.Time
		expectSave        bool
	}{
		// Сценарий 1: текущий месяц (июль) В ДИАПАЗОНЕ (май-сентябрь)
		{
			name:              "July within May-September",
			repoStartMonth:    5, // Май
			repoEndMonth:      9, // Сентябрь
			mockedCurrentTime: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 2: текущий месяц (октябрь) ВНЕ диапазона (май-сентябрь)
		{
			name:              "October outside May-September",
			repoStartMonth:    5, // Май
			repoEndMonth:      9, // Сентябрь
			mockedCurrentTime: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Сценарий 3: текущий месяц (май) НАЧАЛО диапазона (май-сентябрь)
		{
			name:              "May at start of May-September",
			repoStartMonth:    5, // Май
			repoEndMonth:      9, // Сентябрь
			mockedCurrentTime: time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 4: текущий месяц (сентябрь) КОНЕЦ диапазона (май-сентябрь)
		{
			name:              "September at end of May-September",
			repoStartMonth:    5, // Май
			repoEndMonth:      9, // Сентябрь
			mockedCurrentTime: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 5: текущий месяц (январь) ВНУТРИ диапазона с переходом через год (ноябрь-февраль)
		{
			name:              "January within November-February (wrap-around)",
			repoStartMonth:    11, // Ноябрь
			repoEndMonth:      2,  // Февраль
			mockedCurrentTime: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 6: текущий месяц (ноябрь) НАЧАЛО диапазона с переходом через год (ноябрь-февраль)
		{
			name:              "November at start of November-February (wrap-around)",
			repoStartMonth:    11, // Ноябрь
			repoEndMonth:      2,  // Февраль
			mockedCurrentTime: time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 7: текущий месяц (февраль) КОНЕЦ диапазона с переходом через год (ноябрь-февраль)
		{
			name:              "February at end of November-February (wrap-around)",
			repoStartMonth:    11, // Ноябрь
			repoEndMonth:      2,  // Февраль
			mockedCurrentTime: time.Date(2023, time.February, 28, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Сценарий 8: текущий месяц (март) ВНЕ диапазона с переходом через год (ноябрь-февраль)
		{
			name:              "March outside November-February (wrap-around)",
			repoStartMonth:    11, // Ноябрь
			repoEndMonth:      2,  // Февраль
			mockedCurrentTime: time.Date(2023, time.March, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Дополнительный крайний случай: текущий месяц (апрель) ВНЕ диапазона (май-сентябрь) – до начала
		{
			name:              "April outside May-September (before start)",
			repoStartMonth:    5, // Май
			repoEndMonth:      9, // Сентябрь
			mockedCurrentTime: time.Date(2023, time.April, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Дополнительный крайний случай: текущий месяц (октябрь) ВНЕ диапазона с переходом через год (ноябрь-февраль) – между концом и началом
		{
			name:              "October outside November-February (between end and start)",
			repoStartMonth:    11, // Ноябрь
			repoEndMonth:      2,  // Февраль
			mockedCurrentTime: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saver := &mockSaver{}

			repo := NewRepository(tt.repoStartMonth, tt.repoEndMonth)
			repo.AddStore(saver)

			originalNow := now
			now = func() time.Time { return tt.mockedCurrentTime }
			defer func() { now = originalNow }() // Восстанавливаем оригинальный time.Now

			err := repo.Save(dummyData)
			assert.NoError(t, err, "repo.Save не должен возвращать ошибку в этих тестах")

			assert.Equal(t, tt.expectSave, saver.saveCalled, "Статус вызова Save не совпадает")
		})
	}
}
