package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/kuznetsovin/egts-protocol/cli/receiver/config"
	log "github.com/sirupsen/logrus"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	testConfigPath        = "../../configs/config.test.yaml"
	testLogDirRelative    = "../../logs" // Relative to this test file's location (cli/receiver/)
	testLogFileFromConfig = "logs/app_test.log" // Value from config.test.yaml
)

// TestMain handles setup and teardown for tests.
func TestMain(m *testing.M) {
	// Ensure the log directory is clean before tests, if it exists from a previous failed run
	// This path is relative to the project root if tests are run from there,
	// or relative to cli/receiver if run from the package dir.
	// For `os.RemoveAll` it's safer to use a path relative to the test file.
	absLogDir, _ := filepath.Abs(testLogDirRelative)
	_ = os.RemoveAll(absLogDir) // Clean up before tests

	// Run tests
	code := m.Run()

	// Cleanup: remove the logs directory after tests
	err := os.RemoveAll(absLogDir)
	if err != nil {
		// Use standard library log for test cleanup issues, not logrus
		println("WARN: Failed to remove log directory " + absLogDir + ": " + err.Error())
	}

	os.Exit(code)
}

// setupTestLogger initializes logrus with lumberjack based on the provided config.
// It adjusts the log file path to be relative to the project root for test predictability.
func setupTestLogger(t *testing.T, cfg config.Settings) (*lumberjack.Logger, func()) {
	if cfg.LogFilePath == "" {
		log.SetOutput(io.Discard) // Don't log to stdout during tests unless specified by LogFilePath
		return nil, func() {}
	}

	// cfg.LogFilePath is "logs/app_test.log"
	// Tests run in "cli/receiver", so the path needs to be relative to project root.
	// The path from project root will be cfg.LogFilePath itself.
	// We need to make sure that the test can create this file.
	// testLogDirRelative is "../../logs" which correctly points to project_root/logs
	
	// Construct the absolute path for the log file to ensure correctness
	absProjectRoot, err := filepath.Abs("../..")
	if err != nil {
		t.Fatalf("Failed to get absolute path for project root: %v", err)
	}
	lumberjackLogFilePath := filepath.Join(absProjectRoot, cfg.LogFilePath)

	logFileDir := filepath.Dir(lumberjackLogFilePath)
	if _, err := os.Stat(logFileDir); os.IsNotExist(err) {
		if err := os.MkdirAll(logFileDir, 0755); err != nil { // Use 0755 for directory permissions
			t.Fatalf("Failed to create log directory %s: %v", logFileDir, err)
		}
	}
	
	logger := &lumberjack.Logger{
		Filename:   lumberjackLogFilePath,
		MaxSize:    1, // megabytes, small for testing
		MaxBackups: 1, // Keep 1 backup
		MaxAge:     cfg.LogMaxAgeDays,
		Compress:   false, // No compression for easier inspection in tests
	}

	// Output to both stdout (for test visibility if needed) and the lumberjack logger
	// For cleaner test logs, consider only logging to lumberjack and then reading the file.
	// mw := io.MultiWriter(os.Stdout, logger)
	log.SetOutput(logger) // Send logrus output to lumberjack
	log.SetLevel(cfg.GetLogLevel()) // Set level from config

	// Return the logger instance and a cleanup function to close the logger
	cleanupFunc := func() {
		// It's important to close the logger to flush writes to disk,
		// especially before trying to read the file in tests.
		if err := logger.Close(); err != nil {
			t.Logf("Failed to close lumberjack logger: %v", err)
		}
	}

	return logger, cleanupFunc
}

func TestLogFileCreationAndContent(t *testing.T) {
	cfg, err := config.New(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load test config '%s': %v", testConfigPath, err)
	}

	if cfg.LogFilePath == "" {
		t.Skip("LogFilePath is not set in config, skipping log file creation test.")
		return
	}
	if cfg.LogFilePath != testLogFileFromConfig {
        t.Fatalf("Expected LogFilePath '%s' in config, but got '%s'", testLogFileFromConfig, cfg.LogFilePath)
    }


	logger, cleanup := setupTestLogger(t, cfg)
	defer cleanup() // Ensure logger is closed and flushed

	logMessage := "UNIQUE_TEST_MESSAGE_LOG_CREATION_" + time.Now().Format(time.RFC3339Nano)
	log.Infof(logMessage) // Use logrus to log the message

	// Give a brief moment for the log to be written, though Close should handle flushing.
	// time.Sleep(100 * time.Millisecond)


	absProjectRoot, _ := filepath.Abs("../..")
	expectedLogFilePath := filepath.Join(absProjectRoot, cfg.LogFilePath)

	if _, err := os.Stat(expectedLogFilePath); os.IsNotExist(err) {
		t.Fatalf("Log file was not created at %s", expectedLogFilePath)
	}

	content, err := os.ReadFile(expectedLogFilePath)
	if err != nil {
		t.Fatalf("Failed to read log file %s: %v", expectedLogFilePath, err)
	}

	if !strings.Contains(string(content), logMessage) {
		t.Errorf("Log message '%s' not found in log file '%s'. Content:\n%s", logMessage, expectedLogFilePath, string(content))
	}
	
	// Verify that the logger is actually using the file
    if ljInfo, err := os.Stat(logger.Filename); err != nil || ljInfo.Size() == 0 {
         t.Errorf("Lumberjack logger Filename %s does not exist or is empty.", logger.Filename)
    }
}

func TestLogRotationSetting(t *testing.T) {
	cfg, err := config.New(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load test config '%s': %v", testConfigPath, err)
	}

	if cfg.LogFilePath == "" {
		t.Skip("LogFilePath is not set in config, skipping rotation settings test.")
		return
	}
	if cfg.LogMaxAgeDays == 0 {
        t.Log("LogMaxAgeDays is 0 in config, which means logs are not deleted by age. This is a valid setting.")
    }


	logger, cleanup := setupTestLogger(t, cfg)
	defer cleanup()

	if logger == nil {
		// This case implies LogFilePath was empty, which is checked above.
		t.Fatal("Logger was not initialized, but LogFilePath was set.")
	}

	if logger.MaxAge != cfg.LogMaxAgeDays {
		t.Errorf("Lumberjack MaxAge is not set correctly: expected %d from config, got %d in logger", cfg.LogMaxAgeDays, logger.MaxAge)
	}
}

func TestLogDirectoryCreation(t *testing.T) {
	cfg, err := config.New(testConfigPath)
	if err != nil {
		t.Fatalf("Failed to load test config: %v", err)
	}

	// Temporarily use a unique log path to ensure the directory creation is tested
	originalLogPath := cfg.LogFilePath
	uniqueLogSubDir := "unique_test_dir_" + time.Now().Format(time.RFC3339Nano)
	cfg.LogFilePath = filepath.Join("logs", uniqueLogSubDir, "test_app.log")
	t.Logf("Testing with temporary log path: %s", cfg.LogFilePath)
	
	absProjectRoot, _ := filepath.Abs("../..")
	expectedLogDir := filepath.Dir(filepath.Join(absProjectRoot, cfg.LogFilePath))

	// Clean up this specific directory after the test
	defer os.RemoveAll(filepath.Dir(expectedLogDir)) // remove unique_test_dir_... and its parent "logs" if it was created by this test only. Careful with parallel tests.
                                                    // A safer approach is to remove expectedLogDir
    defer os.RemoveAll(expectedLogDir)


	if _, err := os.Stat(expectedLogDir); !os.IsNotExist(err) {
		t.Fatalf("Log directory %s already exists before logger setup", expectedLogDir)
	}

	_, cleanup := setupTestLogger(t, cfg)
	defer cleanup()

	if _, err := os.Stat(expectedLogDir); os.IsNotExist(err) {
		t.Fatalf("Log directory %s was not created by setupTestLogger", expectedLogDir)
	}
	
	// Restore original log path if other tests depend on it (though cfg is scoped here)
	cfg.LogFilePath = originalLogPath
}
