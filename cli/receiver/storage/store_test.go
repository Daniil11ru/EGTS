package storage

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
)

// mockSaver implements the Saver interface for testing.
type mockSaver struct {
	saveCalled bool
	// saveCallCount int // Alternative: count calls
}

// Save marks saveCalled as true.
func (ms *mockSaver) Save(data interface{ ToBytes() ([]byte, error) }) error {
	ms.saveCalled = true
	// ms.saveCallCount++
	return nil
}

// testData is a simple struct for testing the Save method.
type testData struct{}

// ToBytes returns a dummy byte slice and no error.
func (td testData) ToBytes() ([]byte, error) {
	return []byte("test"), nil
}

func TestRepository_Save_DateLogic(t *testing.T) {
	// Discard logs during tests to keep output clean
	log.SetOutput(ioutil.Discard)

	dummyData := testData{}

	tests := []struct {
		name              string
		repoStartMonth    int
		repoEndMonth      int
		mockedCurrentTime time.Time
		expectSave        bool
	}{
		// Scenario 1: Current month (July) WITHIN range (May-September)
		{
			name:              "July within May-September",
			repoStartMonth:    5,  // May
			repoEndMonth:      9,  // September
			mockedCurrentTime: time.Date(2023, time.July, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 2: Current month (October) OUTSIDE range (May-September)
		{
			name:              "October outside May-September",
			repoStartMonth:    5,  // May
			repoEndMonth:      9,  // September
			mockedCurrentTime: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Scenario 3: Current month (May) at START of range (May-September)
		{
			name:              "May at start of May-September",
			repoStartMonth:    5,  // May
			repoEndMonth:      9,  // September
			mockedCurrentTime: time.Date(2023, time.May, 1, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 4: Current month (September) at END of range (May-September)
		{
			name:              "September at end of May-September",
			repoStartMonth:    5,  // May
			repoEndMonth:      9,  // September
			mockedCurrentTime: time.Date(2023, time.September, 30, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 5: Current month (January) WITHIN wrap-around range (November-February)
		{
			name:              "January within November-February (wrap-around)",
			repoStartMonth:    11, // November
			repoEndMonth:      2,  // February
			mockedCurrentTime: time.Date(2023, time.January, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 6: Current month (November) at START of wrap-around range (November-February)
		{
			name:              "November at start of November-February (wrap-around)",
			repoStartMonth:    11, // November
			repoEndMonth:      2,  // February
			mockedCurrentTime: time.Date(2023, time.November, 1, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 7: Current month (February) at END of wrap-around range (November-February)
		{
			name:              "February at end of November-February (wrap-around)",
			repoStartMonth:    11, // November
			repoEndMonth:      2,  // February
			mockedCurrentTime: time.Date(2023, time.February, 28, 0, 0, 0, 0, time.UTC),
			expectSave:        true,
		},
		// Scenario 8: Current month (March) OUTSIDE wrap-around range (November-February)
		{
			name:              "March outside November-February (wrap-around)",
			repoStartMonth:    11, // November
			repoEndMonth:      2,  // February
			mockedCurrentTime: time.Date(2023, time.March, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Additional edge case: Current month (April) OUTSIDE range (May-September) - before start
		{
			name:              "April outside May-September (before start)",
			repoStartMonth:    5,  // May
			repoEndMonth:      9,  // September
			mockedCurrentTime: time.Date(2023, time.April, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
		// Additional edge case: Current month (October) OUTSIDE wrap-around range (November-February) - between end and start
		{
			name:              "October outside November-February (between end and start)",
			repoStartMonth:    11, // November
			repoEndMonth:      2,  // February
			mockedCurrentTime: time.Date(2023, time.October, 15, 0, 0, 0, 0, time.UTC),
			expectSave:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Setup mock saver
			saver := &mockSaver{}

			// Setup repository with the specified month range
			repo := NewRepository(tt.repoStartMonth, tt.repoEndMonth)
			repo.AddStore(saver)

			// Mock time.Now()
			originalNow := now
			now = func() time.Time { return tt.mockedCurrentTime }
			defer func() { now = originalNow }() // Restore original time.Now

			// Call Save
			err := repo.Save(dummyData)
			assert.NoError(t, err, "repo.Save should not return an error in these test cases")

			// Assert
			assert.Equal(t, tt.expectSave, saver.saveCalled, "Save called status mismatch")
		})
	}
}
