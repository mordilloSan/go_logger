# go_logger

![CI](https://github.com/mordilloSan/go_logger/actions/workflows/ci.yml/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/mordilloSan/go_logger/logger.svg)](https://pkg.go.dev/github.com/mordilloSan/go_logger/logger)

Simple, Linux-focused Go logger with automatic caller tagging and systemd-journald integration.

## Features

- **Colorized development output** - INFO/WARN/ERROR to stdout with colors
- **Optional DEBUG** in development via `verbose` flag
- **Production logging** to systemd-journald with `SYSLOG_IDENTIFIER`
- **Automatic caller tagging** - `[package.Function:line]` added to every message
- **Structured logging** - Key-value pairs for better debugging
- **Level filtering** - Control which levels are logged via environment variable

> Note: This package targets Linux due to its journald dependency.

## Install

```bash
go get github.com/mordilloSan/go_logger/logger@latest
```

## Quick Start

### Development Mode

```go
package main

import (
    "time"
    logx "github.com/mordilloSan/go_logger/logger"
)

func main() {
    // Development: INFO/WARN/ERROR enabled, DEBUG only when verbose=true
    logx.Init("development", true)

    logx.Debugf("starting at %v", time.Now())
    logx.Infof("hello %s", "world")
    logx.Warnln("be careful")
    logx.Errorf("oops: %v", "something happened")

    // Structured logging with key-value pairs
    logx.InfoKV("request completed",
        "duration_ms", 42,
        "status", 200,
        "path", "/api/users")
}
```

### Production Mode

```go
// In production, when journald is available, logs go to the systemd journal.
// The SYSLOG_IDENTIFIER is set to the program (binary) name.
logx.Init("production", false)

logx.Infof("server started on port %d", 8080)
logx.ErrorKV("database connection failed",
    "host", "localhost",
    "port", 5432,
    "error", err)
```

Behavior summary:

- **Production + journald available:** Send all levels to journald
- **Production + no journald:** Log plainly to stdout/stderr (INFO/DEBUG to stdout; WARN/ERROR to stderr)
- **Development:** Colorized output to stdout; DEBUG enabled by the `verbose` flag

## API

### Initialization

- `Init(mode string, verbose bool)` - Setup logger for `"development"` or `"production"`

### Formatted Logging (with fmt.Sprintf)

- `Debugf(format string, v ...interface{})`
- `Infof(format string, v ...interface{})`
- `Warnf(format string, v ...interface{})`
- `Errorf(format string, v ...interface{})`
- `Fatalf(format string, v ...interface{})` - Logs and calls `os.Exit(1)`

### Plain Logging (Println-style)

- `Debugln(v ...interface{})`
- `Infoln(v ...interface{})`
- `Warnln(v ...interface{})`
- `Errorln(v ...interface{})`
- `Fatalln(v ...interface{})` - Logs and calls `os.Exit(1)`

### Structured Logging (Key-Value Pairs)

- `DebugKV(msg string, keyvals ...any)`
- `InfoKV(msg string, keyvals ...any)`
- `WarnKV(msg string, keyvals ...any)`
- `ErrorKV(msg string, keyvals ...any)`
- `FatalKV(msg string, keyvals ...any)` - Logs and calls `os.Exit(1)`

Example:
```go
logx.InfoKV("user logged in",
    "user_id", 123,
    "ip", "192.168.1.1",
    "device", "mobile")
```

### API Logging (HTTP Status Code Based)

- `Api(statusCode int, msg string)` - Automatic level selection

Automatically selects log level based on HTTP status code:
- **1xx, 2xx, 3xx** → INFO (green) - Success and redirects
- **4xx** → WARN (yellow) - Client errors
- **5xx** → ERROR (red) - Server errors

Example:
```go
logx.Api(200, "request successful")
logx.Api(404, "resource not found")
logx.Api(500, "internal server error")
```

## Level Filtering

Control which log levels are enabled via the `LOGGER_LEVELS` environment variable:

```bash
# Only log INFO and ERROR
LOGGER_LEVELS="INFO,ERROR" ./myapp

# Only log ERRORS
LOGGER_LEVELS="ERROR" ./myapp

# Log everything (default if not set)
./myapp
```

Valid level names: `DEBUG`, `INFO`, `WARN`, `WARNING`, `ERROR`, `FATAL`

## Output Examples

### Development Mode

```
[INFO] [myapp] 2025/10/25 10:30:45 [main.main:15] server starting on port 8080
[DEBUG] [myapp] 2025/10/25 10:30:45 [main.initDB:23] connecting to database host=localhost port=5432
[INFO] [myapp] 2025/10/25 10:30:46 [main.handleRequest:42] request completed duration_ms=42 status=200 path=/api/users
[ERROR] [myapp] 2025/10/25 10:30:47 [main.processJob:67] job failed job_id=123 error="timeout exceeded"
```

### Production Mode (journald)

Logs go to systemd journal and can be viewed with:

```bash
# View all logs for your app
journalctl -t myapp

# Filter by priority
journalctl -t myapp -p err

# Follow logs in real-time
journalctl -t myapp -f
```

## Use Cases

Perfect for:
- System utilities and daemons
- Web servers and APIs
- CLI applications
- System management dashboards
- Bridge processes requiring elevated privileges

Not ideal for:
- Cloud-native applications (use structured JSON loggers)
- Windows/macOS applications (journald dependency)
- Microservices sending logs to centralized systems

## Compatibility

- **Go:** 1.22+
- **OS:** Linux (journald integration)

## Testing

Run all tests:

```bash
go test ./...
go test -v ./...  # verbose output
```

What the tests verify:
- Production fallback logs to stdout/stderr when journald is unavailable
- Journald path sends messages with correct priority and identifier
- Development mode toggles DEBUG via the `verbose` flag
- Level filtering works correctly

Tests do not require a running journald; the logger uses injection points during tests.

## Project Layout

```
go_logger/
├── main.go              # Example app
├── logger/
│   ├── logger.go        # Core implementation
│   ├── doc.go          # Package documentation
│   └── *_test.go       # Tests
├── go.mod
└── README.md
```

Run the example app:

```bash
go run .                # development mode
go run . production     # production mode
```

## Common Tasks

### Using Makefile (Recommended)

```bash
make              # Run fmt, vet, and test (default)
make test         # Run all tests with verbose output
make fmt          # Format code
make vet          # Run static analysis
make pre-release  # Run all checks before creating a release
make clean        # Clean build cache
make help         # Show all available targets
```

### Using Go Commands Directly

```bash
go fmt ./...      # Format code
go vet ./...      # Lint
go test ./...     # Run tests
go test -v ./...  # Run tests with verbose output
```

## Why This Logger?

- **Simple:** Single `Init()` call, no configuration structs
- **Zero dependencies:** Just the Go standard library + journald
- **Automatic caller info:** No manual tagging needed
- **Production-ready:** Integrates with systemd-journald
- **Structured logging:** Key-value pairs for better debugging

## License

MIT. See `LICENSE`.
