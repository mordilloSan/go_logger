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
make test              # Run all 27 tests
go test ./...          # Or use go directly
go test -v ./...       # Verbose output
make test-concurrency  # Demo concurrency with live progress
```

### Test Coverage (27 tests total)

**Concurrency Tests** - Prove thread-safety under extreme load:
- 10,000 goroutines × 100 messages × 4 levels = **4 million log operations**
- 100+ concurrent goroutines using all logging methods
- Real-time progress demo showing mutex effectiveness
- All tests verify **zero garbled output**

**Fatal Method Tests** - Verify logging before process exit:
- Confirms `Fatalf`, `Fatalln`, `FatalKV` write logs before `os.Exit(1)`
- Tests level filtering and output formatting
- Uses subprocess execution for proper testing

**Crash Scenario Tests** - Prove log flushing under failure:
- 5,000 rapid log operations all flushed correctly
- Panic recovery with proper log flushing
- Validates v1.1.0 claims about crash resilience

**Core Functionality Tests**:
- Production fallback when journald unavailable
- Journald integration with correct priorities
- Development mode DEBUG toggling
- Level filtering
- Caller info tagging
- Structured logging (KV pairs)

Tests do not require running journald; the logger uses injection points during tests.

### See It In Action

Watch the mutex prevent garbled output from 50 concurrent workers:

```bash
make test-concurrency
```

Output shows clean progress updates:
```
Starting concurrency test: 50 workers × 100 tasks = 5000 total operations
progress completed=1900 total=5000 percent=38.0% active_workers=50 tasks_per_sec=9500
progress completed=3800 total=5000 percent=76.0% active_workers=50 tasks_per_sec=9500
✓ CONCURRENCY TEST COMPLETE!
final stats: 5000 operations in 526ms = 9498 ops/sec - NO GARBLED OUTPUT
```

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
make                   # Run fmt, vet, and test (default)
make test              # Run all tests with verbose output
make test-concurrency  # Demo real-time concurrent logging (100 goroutines)
make fmt               # Format code
make vet               # Run static analysis
make pre-release       # Run all checks before creating a release
make clean             # Clean build cache
make help              # Show all available targets
```

**See the mutex in action:** Run `make test-concurrency` to watch 100 concurrent goroutines logging in real-time with color-coded output and no garbled lines!

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
