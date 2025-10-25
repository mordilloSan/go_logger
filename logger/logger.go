package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"

	"github.com/coreos/go-systemd/v22/journal"
)

// Levels define log severity.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
)

// global state
var (
	// log.Logger instances for formatted output
	Debug   = log.New(io.Discard, "", 0)
	Info    = log.New(io.Discard, "", 0)
	Warning = log.New(io.Discard, "", 0)
	Error   = log.New(io.Discard, "", 0)

	programName string
	mode        string
	verbose     bool

	// enabled levels (for filtering)
	enabledLevels = map[Level]bool{
		DebugLevel: true,
		InfoLevel:  true,
		WarnLevel:  true,
		ErrorLevel: true,
	}
)

// Dependency injection points for testing journald behavior and outputs.
var (
	journalIsEnabled = journal.Enabled
	journalSendFunc  = func(msg string, priority journal.Priority, vars map[string]string) error {
		return journal.Send(msg, priority, vars)
	}
	outStdout io.Writer = os.Stdout
	outStderr io.Writer = os.Stderr
)

// Init initializes the logger for development or production mode.
// Development uses colored stdout; production prefers journald, else stdout/stderr.
// Set verbose=true to enable DEBUG logs in development mode.
// Respects LOGGER_LEVELS environment variable for filtering (e.g., "INFO,ERROR").
func Init(logMode string, verboseMode bool) {
	programName = filepath.Base(os.Args[0])
	mode = logMode
	verbose = verboseMode

	// Parse level filtering from environment
	if levels := os.Getenv("LOGGER_LEVELS"); levels != "" {
		enabledLevels = parseLevels(levels)
	}

	if mode == "production" {
		if journalIsEnabled() {
			Debug = log.New(journalWriter{journal.PriDebug}, "", 0)
			Info = log.New(journalWriter{journal.PriInfo}, "", 0)
			Warning = log.New(journalWriter{journal.PriWarning}, "", 0)
			Error = log.New(journalWriter{journal.PriErr}, "", 0)
		} else {
			Debug = newPlainLogger(outStdout, "DEBUG")
			Info = newPlainLogger(outStdout, "INFO")
			Warning = newPlainLogger(outStderr, "WARN")
			Error = newPlainLogger(outStderr, "ERROR")
		}
		return
	}

	// Development mode
	Debug = newDevLogger(outStdout, "DEBUG", verboseMode)
	Info = newDevLogger(outStdout, "INFO", true)
	Warning = newDevLogger(outStdout, "WARN", true)
	Error = newDevLogger(outStdout, "ERROR", true)
}

// parseLevels parses a comma-separated list of level names.
// Empty string enables all levels.
func parseLevels(s string) map[Level]bool {
	m := map[Level]bool{}
	s = strings.TrimSpace(s)
	if s == "" {
		m[DebugLevel] = true
		m[InfoLevel] = true
		m[WarnLevel] = true
		m[ErrorLevel] = true
		return m
	}
	for _, p := range strings.Split(s, ",") {
		switch strings.ToUpper(strings.TrimSpace(p)) {
		case "DEBUG":
			m[DebugLevel] = true
		case "INFO":
			m[InfoLevel] = true
		case "WARN", "WARNING":
			m[WarnLevel] = true
		case "ERROR":
			m[ErrorLevel] = true
		}
	}
	return m
}

// levelString returns the string representation of a log level.
func levelString(l Level) string {
	switch l {
	case DebugLevel:
		return "DEBUG"
	case InfoLevel:
		return "INFO"
	case WarnLevel:
		return "WARN"
	case ErrorLevel:
		return "ERROR"
	default:
		return "INFO"
	}
}

// isLevelEnabled checks if a level is enabled for logging.
func isLevelEnabled(level Level) bool {
	return enabledLevels[level]
}

// journalWriter writes to systemd journal with the program name as identifier.
type journalWriter struct {
	priority journal.Priority
}

func (j journalWriter) Write(p []byte) (int, error) {
	msg := strings.TrimSuffix(string(p), "\n")
	err := journalSendFunc(msg, j.priority, map[string]string{
		"SYSLOG_IDENTIFIER": programName,
	})
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// newDevLogger returns a colored logger for the level, or discards if disabled.
func newDevLogger(out io.Writer, level string, enabled bool) *log.Logger {
	if !enabled {
		return log.New(io.Discard, "", 0)
	}
	colors := map[string]string{
		"DEBUG": "\033[36m",
		"INFO":  "\033[32m",
		"WARN":  "\033[33m",
		"ERROR": "\033[31m",
	}
	reset := "\033[0m"
	levelLabel := fmt.Sprintf("%s[%s]%s", colors[level], level, reset)
	prefix := fmt.Sprintf("%s [%s] ", levelLabel, programName)
	return log.New(out, prefix, log.LstdFlags)
}

// newPlainLogger returns a non-colored logger for production stdout/stderr fallback.
func newPlainLogger(out io.Writer, level string) *log.Logger {
	prefix := fmt.Sprintf("[%s] [%s] ", level, programName)
	return log.New(out, prefix, log.LstdFlags)
}

// getCallerInfo returns formatted caller information at the specified stack depth.
// Returns "package.Function" format for better log clarity.
func getCallerInfo(depth int) string {
	pc, _, line, ok := runtime.Caller(depth)
	if !ok {
		return "unknown"
	}
	fn := runtime.FuncForPC(pc)
	if fn == nil {
		return "unknown"
	}
	full := fn.Name()
	// Strip package path, keep package.Function
	lastSlash := strings.LastIndex(full, "/")
	if lastSlash >= 0 && lastSlash+1 < len(full) {
		full = full[lastSlash+1:]
	}
	return fmt.Sprintf("%s:%d", full, line)
}

// encodeFields formats key-value pairs as "key=value" strings.
func encodeFields(keyvals ...any) string {
	if len(keyvals) == 0 {
		return ""
	}
	parts := make([]string, 0, len(keyvals)/2)
	for i := 0; i+1 < len(keyvals); i += 2 {
		key, ok := keyvals[i].(string)
		if !ok {
			continue
		}
		parts = append(parts, fmt.Sprintf("%s=%v", key, keyvals[i+1]))
	}
	if len(parts) == 0 {
		return ""
	}
	return " " + strings.Join(parts, " ")
}

// --- Structured logging methods (key-value pairs) ---

// DebugKV logs a debug message with structured key-value pairs.
// The caller function name and line number are automatically included.
func DebugKV(msg string, keyvals ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Debug.Printf("[%s] %s%s", caller, msg, fields)
}

// InfoKV logs an info message with structured key-value pairs.
// The caller function name and line number are automatically included.
func InfoKV(msg string, keyvals ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Info.Printf("[%s] %s%s", caller, msg, fields)
}

// WarnKV logs a warning message with structured key-value pairs.
// The caller function name and line number are automatically included.
func WarnKV(msg string, keyvals ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Warning.Printf("[%s] %s%s", caller, msg, fields)
}

// ErrorKV logs an error message with structured key-value pairs.
// The caller function name and line number are automatically included.
func ErrorKV(msg string, keyvals ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Error.Printf("[%s] %s%s", caller, msg, fields)
}

// --- Formatted logging methods ---

// Debugf logs a debug message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
func Debugf(format string, v ...interface{}) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Debug.Println(msg)
}

// Infof logs an informational message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
func Infof(format string, v ...interface{}) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Info.Println(msg)
}

// Warnf logs a warning message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
func Warnf(format string, v ...interface{}) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Warning.Println(msg)
}

// Errorf logs an error message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
func Errorf(format string, v ...interface{}) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Error.Println(msg)
}

// --- Plain "Println" helpers for literal messages ---

// Debugln logs a debug message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
func Debugln(v ...interface{}) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Debug.Println(msg)
}

// Infoln logs an informational message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
func Infoln(v ...interface{}) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Info.Println(msg)
}

// Warnln logs a warning message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
func Warnln(v ...interface{}) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Warning.Println(msg)
}

// Errorln logs an error message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
func Errorln(v ...interface{}) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Error.Println(msg)
}
