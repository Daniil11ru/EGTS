package config

import (
	"github.com/stretchr/testify/assert"
	"io/ioutil"
	"os"
	"testing"

	log "github.com/sirupsen/logrus"
)

func TestConfigLoad(t *testing.T) {
	// To prevent log output during tests
	log.SetOutput(ioutil.Discard)

	cfg := `host: "127.0.0.1"
port: "5020"
conn_ttl: 10
log_level: "DEBUG"

storage:
  rabbitmq:
    host: "localhost"
    port: "5672"
    user: "guest"
    password: "guest"
    exchange: "receiver"
  postgresql:
    host: "localhost"
    port: "5432"
    user: "postgres"
    password: "postgres"
    database: "receiver"
    table: "points"
    sslmode: "disable"
`

	file, err := ioutil.TempFile("/tmp", "config.toml")
	if !assert.NoError(t, err) {
		return
	}
	defer os.Remove(file.Name())

	if _, err = file.WriteString(cfg); !assert.NoError(t, err) {
		return
	}

	conf, err := New(file.Name())
	if assert.NoError(t, err) {
		assert.Equal(t, Settings{
			Host:     "127.0.0.1",
			Port:     "5020",
			ConnTTl:  10,
			LogLevel: "DEBUG",
			// LogFilePath and LogMaxAgeDays remain as zero values if not in YAML
			Store: map[string]map[string]string{
				"postgresql": {
					"host":     "localhost",
					"port":     "5432",
					"user":     "postgres",
					"password": "postgres",
					"database": "receiver",
					"table":    "points",
					"sslmode":  "disable",
				},
				"rabbitmq": {
					"exchange": "receiver",
					"host":     "localhost",
					"password": "guest",
					"port":     "5672",
					"user":     "guest",
				},
			},
			DBSaveMonthStart: 5, // Default value
			DBSaveMonthEnd:   9, // Default value
		},
			conf,
		)
	}
}

func TestDBSaveMonthRange(t *testing.T) {
	// To prevent log output during tests
	log.SetOutput(ioutil.Discard)

	tests := []struct {
		name                 string
		yamlContent          string
		expectedStartMonth   int
		expectedEndMonth     int
		expectError          bool
		deleteFileAfterwards bool
	}{
		{
			name: "Fields provided in YAML",
			yamlContent: `
db_save_month_start: 3
db_save_month_end: 10
`,
			expectedStartMonth: 3,
			expectedEndMonth:   10,
			deleteFileAfterwards: true,
		},
		{
			name: "Fields not provided (defaults)",
			yamlContent: `
# empty config
`,
			expectedStartMonth: 5, // Default May
			expectedEndMonth:   9, // Default September
			deleteFileAfterwards: true,
		},
		{
			name: "Invalid month values (0, 13)",
			yamlContent: `
db_save_month_start: 0
db_save_month_end: 13
`,
			expectedStartMonth: 5, // Default May (due to validation)
			expectedEndMonth:   9, // Default September (due to validation)
			deleteFileAfterwards: true,
		},
		{
			name: "Start month after end month (invalid simple range)",
			yamlContent: `
db_save_month_start: 11
db_save_month_end: 2
`,
			expectedStartMonth: 5, // Default May (due to validation)
			expectedEndMonth:   9, // Default September (due to validation)
			deleteFileAfterwards: true,
		},
		{
			name:                 "Non-existent config file",
			yamlContent:          "", // No content, file won't be created for this specific sub-test
			expectedStartMonth:   0,  // Expect 0 as defaults are not applied if file read fails
			expectedEndMonth:     0,  // Expect 0 as defaults are not applied if file read fails
			expectError:          true, // Expect error when loading non-existent file
			deleteFileAfterwards: false, // No file to delete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var confPath string
			var err error

			if tt.yamlContent != "" || tt.name == "Fields not provided (defaults)" { // Second condition to create empty file for default test
				file, err := ioutil.TempFile("", "test_config_*.yaml")
				if !assert.NoError(t, err, "Failed to create temp file for test: %s", tt.name) {
					return
				}
				confPath = file.Name()

				if tt.yamlContent != "" {
					_, err = file.WriteString(tt.yamlContent)
					if !assert.NoError(t, err, "Failed to write to temp file for test: %s", tt.name) {
						file.Close()
						if tt.deleteFileAfterwards { os.Remove(confPath) }
						return
					}
				}
				file.Close() // Close the file before New() tries to read it
			} else if tt.name == "Non-existent config file" {
				confPath = "/tmp/non_existent_config_for_test.yaml" // A path that should not exist
			}


			cfg, err := New(confPath)

			if tt.deleteFileAfterwards && confPath != "" {
				errRemove := os.Remove(confPath)
				assert.NoError(t, errRemove, "Failed to remove temp file: %s for test: %s", confPath, tt.name)
			}

			if tt.expectError {
				assert.Error(t, err, "Expected an error for test: %s", tt.name)
				// When an error is expected (like file not found),
				// the cfg object might not be fully initialized or might be zeroed.
				// When an error is expected (like file not found),
				// the cfg object is the zero-value one returned before default logic is applied.
				assert.Equal(t, tt.expectedStartMonth, cfg.DBSaveMonthStart, "Start month should be 0 for error test: %s", tt.name)
				assert.Equal(t, tt.expectedEndMonth, cfg.DBSaveMonthEnd, "End month should be 0 for error test: %s", tt.name)
				// Also check getters which would return these 0 values
				assert.Equal(t, tt.expectedStartMonth, cfg.GetDBSaveMonthStart(), "Getter start month mismatch for error test: %s", tt.name)
				assert.Equal(t, tt.expectedEndMonth, cfg.GetDBSaveMonthEnd(), "Getter end month mismatch for error test: %s", tt.name)
			} else {
				if !assert.NoError(t, err, "Unexpected error for test: %s", tt.name) {
					return
				}
				assert.Equal(t, tt.expectedStartMonth, cfg.GetDBSaveMonthStart(), "Start month mismatch for test: %s", tt.name)
				assert.Equal(t, tt.expectedEndMonth, cfg.GetDBSaveMonthEnd(), "End month mismatch for test: %s", tt.name)
			}
		})
	}
}
