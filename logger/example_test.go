package logger_test

import (
	"github.com/mordilloSan/go_logger/logger"
)

// This example shows basic development usage with DEBUG enabled.
func ExampleInit_development() {
	logger.Init("development", true)
	logger.Debugf("debug is on")
	logger.Infof("hello %s", "world")
	logger.Warnln("be careful")
	logger.Errorf("oops: %v", "boom")
}

// This example shows a typical production setup.
func ExampleInit_production() {
	// When journald is available, messages are sent to the systemd journal.
	// Otherwise, logs go to stdout/stderr plainly.
	logger.Init("production", false)
	logger.Infof("ready")
}

// This example demonstrates structured logging with key-value pairs.
func ExampleInfoKV() {
	logger.Init("development", true)

	// Structured logging is great for debugging and analysis
	logger.InfoKV("request completed",
		"duration_ms", 42,
		"status", 200,
		"path", "/api/users",
		"method", "GET")

	logger.ErrorKV("database connection failed",
		"host", "localhost",
		"port", 5432,
		"retry_count", 3)
}

// This example shows how to filter log levels via environment variable.
func ExampleInit_levelFiltering() {
	// Set LOGGER_LEVELS="INFO,ERROR" before running to disable DEBUG and WARN
	logger.Init("development", true)

	logger.Debugf("this won't appear if DEBUG is filtered")
	logger.Infof("this will appear")
	logger.Warnf("this won't appear if WARN is filtered")
	logger.Errorf("this will appear")
}
