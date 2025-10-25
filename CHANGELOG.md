# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [v1.1.0] - 2025-10-25

### Added

- **Fatal logging methods** - New methods that log and exit:
  - `Fatalf(format string, v ...any)` - Formatted fatal logging
  - `Fatalln(v ...any)` - Plain fatal logging
  - `FatalKV(msg string, keyvals ...any)` - Structured fatal logging
  - All fatal methods call `os.Exit(1)` after logging
  - Fatal messages use magenta color in development mode
  - Fatal messages use `journal.PriCrit` in production with journald

- **API logging method** - HTTP status code-based logging:
  - `Api(statusCode int, msg string)` - Automatic level selection based on status code
  - 1xx, 2xx, 3xx → INFO (green)
  - 4xx → WARN (yellow)
  - 5xx → ERROR (red)
  - Includes status code in output: `[200] api call successful`

- **Thread-safety** - All logging methods now use `sync.Mutex`:
  - Prevents mixed/garbled log output from concurrent goroutines
  - Safe for high-concurrency applications (web servers, APIs)
  - Tested with 100+ concurrent goroutines

### Changed

- **Removed program name from output** - Cleaner logs without redundant `[backend]` tag:
  - Before: `[INFO] [backend] 2025/10/25 22:57:34 [file.Function:158] message`
  - After: `[INFO] 2025/10/25 22:57:34 [file.Function:158] message`
  - Caller info `[file.Function:line]` is sufficient for debugging

- **Organized functions by functionality** instead of log level:
  - Formatted methods: `Debugf`, `Infof`, `Warnf`, `Errorf`, `Fatalf`
  - Plain methods: `Debugln`, `Infoln`, `Warnln`, `Errorln`, `Fatalln`
  - Structured methods: `DebugKV`, `InfoKV`, `WarnKV`, `ErrorKV`, `FatalKV`
  - API method: `Api`

- **Improved API status code handling**:
  - 3xx redirect codes now map to INFO instead of WARN (correct behavior)
  - Better alignment with HTTP semantics

- **Updated level filtering** to support `FATAL` in `LOGGER_LEVELS` environment variable

### Performance

- Thread-safe logging with minimal overhead
- Mutex ensures atomic log writes (no garbled output)
- Logs always flushed even on process crashes (`defer` ensures cleanup)

## [v1.0.0] - 2025-10-25

### Initial Release

Simple, Linux-focused logger optimized for system utilities and single-deployment applications.

#### Features

- **Simple API** - Single `Init(mode, verbose)` call, no complex configuration
- **Global package-level functions** - No dependency injection needed
- **Automatic caller tagging** - `[package.Function:line]` in every log message
- **Structured logging** - Key-value pairs with `*KV()` methods:
  - `DebugKV(msg string, keyvals ...any)`
  - `InfoKV(msg string, keyvals ...any)`
  - `WarnKV(msg string, keyvals ...any)`
  - `ErrorKV(msg string, keyvals ...any)`
- **Level filtering** - Via `LOGGER_LEVELS` environment variable
- **Journald integration** - First-class systemd-journald support with `SYSLOG_IDENTIFIER`
- **Development mode** - Colorized output (cyan DEBUG, green INFO, yellow WARN, red ERROR)
- **Production mode** - Logs to journald or stdout/stderr fallback

#### API

Initialization:
- `Init(mode string, verbose bool)`

Formatted logging:
- `Debugf(format string, v ...interface{})`
- `Infof(format string, v ...interface{})`
- `Warnf(format string, v ...interface{})`
- `Errorf(format string, v ...interface{})`

Plain logging:
- `Debugln(v ...interface{})`
- `Infoln(v ...interface{})`
- `Warnln(v ...interface{})`
- `Errorln(v ...interface{})`

Structured logging (NEW):
- `DebugKV(msg string, keyvals ...any)`
- `InfoKV(msg string, keyvals ...any)`
- `WarnKV(msg string, keyvals ...any)`
- `ErrorKV(msg string, keyvals ...any)`

#### Perfect For

- System utilities and daemons
- Web servers and APIs
- CLI applications
- System management dashboards
- Bridge processes requiring elevated privileges

#### Not Ideal For

- Cloud-native microservices (use structured JSON loggers)
- Windows/macOS applications (no journald)
- Distributed tracing (use OpenTelemetry)

#### Requirements

- Go 1.22+
- Linux (for journald integration)
- `github.com/coreos/go-systemd/v22`

[v1.1.0]: https://github.com/mordilloSan/go_logger/releases/tag/v1.1.0
[v1.0.0]: https://github.com/mordilloSan/go_logger/releases/tag/v1.0.0
