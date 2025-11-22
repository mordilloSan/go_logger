// Package logger provides a simple leveled logger with
// automatic caller function tagging and optional file output.
//
// # Development Mode
//
// Colorized INFO/WARN/ERROR to stdout, DEBUG enabled only when verbose is true.
//
// # Features
//
//   - Global package-level functions (no dependency injection needed)
//   - Automatic caller tagging [package.Function:line]
//   - Structured logging with key-value pairs
//   - Level filtering via LOGGER_LEVELS environment variable
//   - Optional file logging with color stripping for files
//
// # Usage
//
// Initialize once at startup:
//
//	logger.Init("development", true)  // verbose debug mode
//	logger.Init("production", false)   // production mode
//
// Use formatted logging:
//
//	logger.Infof("server started on port %d", 8080)
//	logger.Errorf("failed to connect: %v", err)
//
// Use structured logging with key-value pairs:
//
//	logger.InfoKV("request completed",
//	    "duration_ms", 42,
//	    "status", 200,
//	    "path", "/api/users")
//
// Fatal logging (logs and exits):
//
//	logger.Fatalf("critical error: %v", err)
//	logger.FatalKV("shutdown required", "reason", "out of memory")
//
// # Level Filtering
//
// Control which levels are logged via environment variable:
//
//	LOGGER_LEVELS="INFO,ERROR" ./myapp
//
// This package is lightweight and has no external dependencies.
package logger
