# Changelog

## v2.0.0 - Simplified for LinuxIO (2025-10-25)

Major refactoring to simplify the logger for single-deployment Linux applications.

### Breaking Changes

- **Removed JSON output** - Journald already provides structured logging
- **Removed Config struct** - Single `Init(mode, verbose)` is sufficient
- **Removed NewLogger/Logger interface** - No dependency injection needed
- **Removed context extraction** - `ContextKeys`, `*Context()` methods removed
- **Removed API helpers** - `API()`, `APIContext()` methods removed
- **Removed `With()` and `WithGroup()`** - Simplified to global functions only

### New Features

- **Structured logging** - New `*KV()` methods for key-value pairs:
  - `InfoKV(msg string, keyvals ...any)`
  - `ErrorKV(msg string, keyvals ...any)`
  - `DebugKV(msg string, keyvals ...any)`
  - `WarnKV(msg string, keyvals ...any)`

- **Level filtering** - Control which levels are logged:
  - Via `LOGGER_LEVELS` environment variable
  - Examples: `LOGGER_LEVELS="INFO,ERROR"`, `LOGGER_LEVELS="ERROR"`
  - Supports: DEBUG, INFO, WARN, WARNING, ERROR

- **Better caller detection** - Now includes line numbers:
  - Format: `[package.Function:line]` instead of `[package.Function]`
  - More accurate stack depth handling

### Improvements

- Reduced code size by ~50% (503 lines → 355 lines)
- Simpler API surface (3 functions → same 3 + 4 new KV variants)
- Better test coverage for new features
- Clearer documentation focused on single-deployment use cases

### Migration Guide

If you were using the old API:

**Before (v1.x):**
```go
cfg := logger.Config{
    Mode:        "production",
    JSON:        true,
    UseJournald: true,
}
log, _ := logger.NewLogger(cfg)
log.Info("server started", "port", 8080)
```

**After (v2.0):**
```go
logger.Init("production", false)
logger.InfoKV("server started", "port", 8080)
```

**Removed features and alternatives:**
- JSON output → Use journald's native structured logging
- Context extraction → Pass values explicitly as key-value pairs
- API helpers → Use regular logging with status codes as KV pairs
- `With()`/`WithGroup()` → Use package-level functions directly

### Why These Changes?

This logger is optimized for:
- System utilities and daemons
- Single-deployment applications
- Linux systems with journald
- Simple, straightforward logging needs

Not ideal for:
- Cloud-native microservices (use structured JSON loggers)
- Complex distributed tracing (use OpenTelemetry)
- Windows/macOS applications (no journald)

## v1.0.0 - Initial Release

Original feature-rich logger with DI, JSON output, and context extraction.
