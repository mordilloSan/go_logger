package main

import (
	"os"
	"time"

	logx "github.com/mordilloSan/go_logger/logger"
)

// Example demonstrating the simplified go_logger usage.
func main() {
	mode := "development"
	if len(os.Args) > 1 {
		mode = os.Args[1]
	}

	// Simple init - just mode and verbose flag
	logx.Init(mode, mode == "development")

	// Formatted logging (classic)
	logx.Debugf("starting at %v", time.Now())
	logx.Infof("hello %s", "world")
	logx.Warnln("be careful")
	logx.Errorf("oops: %v", "something happened")

	// Structured logging with key-value pairs (NEW!)
	logx.InfoKV("request completed",
		"duration_ms", 42,
		"status", 200,
		"path", "/api/users",
		"method", "GET")

	logx.ErrorKV("database connection failed",
		"host", "localhost",
		"port", 5432,
		"retry_count", 3,
		"error", "connection timeout")

	logx.DebugKV("cache lookup",
		"key", "user:123",
		"hit", true,
		"ttl_seconds", 300)

	// API logging (automatic level selection based on HTTP status code)
	logx.Api(200, "request successful")
	logx.Api(301, "redirect to new location")
	logx.Api(404, "resource not found")
	logx.Api(500, "internal server error")

	// Uncomment to test Fatal methods (will exit the program):
	// logx.Fatalf("critical error: %v", "system failure")
	// logx.FatalKV("critical error", "component", "database", "action", "shutdown")
}
