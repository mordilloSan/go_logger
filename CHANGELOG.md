# Changelog

All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

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

[v1.0.0]: https://github.com/mordilloSan/go_logger/releases/tag/v1.0.0
