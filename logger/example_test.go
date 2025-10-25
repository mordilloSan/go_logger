package logger_test

import (
	logx "github.com/mordilloSan/go_logger/logger"
)

// This example shows basic development usage with DEBUG enabled.
func ExampleInit_development() {
	logx.Init("development", true)
	logx.Debugf("debug is on")
	logx.Infof("hello %s", "world")
	logx.Warnln("be careful")
	logx.Errorf("oops: %v", "boom")
}

// This example shows a typical production setup.
func ExampleInit_production() {
	// When journald is available, messages are sent to the systemd journal.
	// Otherwise, logs go to stdout/stderr plainly.
	logx.Init("production", false)
	logx.Infof("ready")
}

// This example demonstrates structured logging with key-value pairs.
func ExampleInfoKV() {
	logx.Init("development", true)

	// Structured logging is great for debugging and analysis
	logx.InfoKV("request completed",
		"duration_ms", 42,
		"status", 200,
		"path", "/api/users",
		"method", "GET")

	logx.ErrorKV("database connection failed",
		"host", "localhost",
		"port", 5432,
		"retry_count", 3)
}

// This example shows how to filter log levels via environment variable.
func ExampleInit_levelFiltering() {
	// Set LOGGER_LEVELS="INFO,ERROR" before running to disable DEBUG and WARN
	logx.Init("development", true)

	logx.Debugf("this won't appear if DEBUG is filtered")
	logx.Infof("this will appear")
	logx.Warnf("this won't appear if WARN is filtered")
	logx.Errorf("this will appear")
}
