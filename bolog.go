package bolog

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/natefinch/lumberjack"
)

// ConfigLogger defines the configuration structure for the logger.
type ConfigLogger struct {
	LogDir     string `json:"logDir"`     // Directory for storing logs
	MaxSize    int    `json:"maxsize"`    // Maximum log file size in megabytes
	MaxBackups int    `json:"maxbackups"` // Maximum number of old log files to retain
	MaxAge     int    `json:"maxage"`     // Maximum number of days to retain old log files
	Compress   bool   `json:"compress"`   // Compress old log files
	Timezone   string `json:"timezone"`   // Timezone
}

// Logger is a wrapper around lumberjack.Logger.
type Logger struct {
	lumberjack.Logger
	config ConfigLogger
}

// InitializeLoggerFromConfig reads a configuration file and initializes a logger.
// It returns a pointer to the initialized Logger or an error if the process fails.
func InitializeLoggerFromConfig(configFile string) (*Logger, error) {
	loggerConfig, err := LoadLoggerConfig(configFile)
	if err != nil {
		return nil, err
	}

	return SetupLogger(loggerConfig), nil
}

// SetupLogger creates the log directory and initializes a lumberjack.Logger with the specified configurations.
// It returns a pointer to the initialized Logger.
func SetupLogger(config ConfigLogger) *Logger {
	err := os.MkdirAll(config.LogDir, os.ModePerm)
	if err != nil {
		log.Fatal(err)
	}

	logPath := filepath.Join(config.LogDir, getLogFileName(config.Timezone))
	return &Logger{
		Logger: lumberjack.Logger{
			Filename:   logPath,
			MaxSize:    config.MaxSize,
			MaxBackups: config.MaxBackups,
			MaxAge:     config.MaxAge,
			Compress:   config.Compress,
		},
		config: config,
	}
}

// Logf logs a formatted message with the current time and timezone from the configuration.
func (l *Logger) Logf(format string, v ...interface{}) {
	currentTime := time.Now().In(getTimezone(l.config.Timezone))
	message := fmt.Sprintf("[%s] -- "+format, append([]interface{}{currentTime.Format("2006-01-02 15:04:05")}, v...)...)

	if _, err := l.Write([]byte(message + "\n")); err != nil {
		log.Printf("Error writing log: %v", err)
	}
}

// getLogFileName generates a log file name based on the current date and timezone.
func getLogFileName(timezone string) string {
	currentTime := time.Now().In(getTimezone(timezone))
	return currentTime.Format("log_20060102.txt")
}

// getTimezone returns a time.Location object for the specified timezone,
// defaulting to UTC if the timezone is invalid.
func getTimezone(timezone string) *time.Location {
	loc, err := time.LoadLocation(timezone)
	if err != nil {
		loc = time.UTC
	}
	return loc
}

// LoadLoggerConfig reads and decodes a JSON configuration file into a ConfigLogger struct.
func LoadLoggerConfig(configPath string) (ConfigLogger, error) {
	file, err := os.Open(configPath)
	if err != nil {
		return ConfigLogger{}, err
	}
	defer func(file *os.File) {
		// Close the file and handle any potential errors
		err := file.Close()
		if err != nil {
			log.Fatal("Error closing the file:", err)
		}
	}(file)

	var config ConfigLogger
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return ConfigLogger{}, err
	}

	return config, nil
}
