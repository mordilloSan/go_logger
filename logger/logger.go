package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"github.com/coreos/go-systemd/v22/journal"
)

// Levels define log severity.
type Level int

const (
	DebugLevel Level = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
)

// global state
var (
	// log.Logger instances for formatted output
	Debug   = log.New(io.Discard, "", 0)
	Info    = log.New(io.Discard, "", 0)
	Warning = log.New(io.Discard, "", 0)
	Error   = log.New(io.Discard, "", 0)
	Fatal   = log.New(io.Discard, "", 0)

	programName string // used for journald SYSLOG_IDENTIFIER

	// Mutex for thread-safe logging across concurrent goroutines
	logMutex sync.Mutex

	// enabled levels (for filtering)
	enabledLevels = map[Level]bool{
		DebugLevel: true,
		InfoLevel:  true,
		WarnLevel:  true,
		ErrorLevel: true,
		FatalLevel: true,
	}

	// logFile holds the file handle for file logging (if enabled)
	logFile *os.File
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
	InitWithFile(logMode, verboseMode, "")
}

// InitWithFile initializes the logger with optional file logging.
// If filePath is non-empty, logs will be written to both console and file.
// The file is created with append mode and 0644 permissions.
// Call Close() to properly close the log file when shutting down.
func InitWithFile(logMode string, verboseMode bool, filePath string) {
	programName = filepath.Base(os.Args[0])

	// Parse level filtering from environment
	if levels := os.Getenv("LOGGER_LEVELS"); levels != "" {
		enabledLevels = parseLevels(levels)
	}

	// Open log file if specified
	var fileWriter io.Writer
	if filePath != "" {
		f, err := os.OpenFile(filePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
		if err != nil {
			fmt.Fprintf(os.Stderr, "failed to open log file %s: %v\n", filePath, err)
		} else {
			logFile = f
			fileWriter = f
		}
	}

	if logMode == "production" {
		if journalIsEnabled() {
			Debug = log.New(journalWriter{journal.PriDebug}, "", 0)
			Info = log.New(journalWriter{journal.PriInfo}, "", 0)
			Warning = log.New(journalWriter{journal.PriWarning}, "", 0)
			Error = log.New(journalWriter{journal.PriErr}, "", 0)
			Fatal = log.New(journalWriter{journal.PriCrit}, "", 0)
		} else {
			Debug = newPlainLogger(outStdout, "DEBUG", fileWriter)
			Info = newPlainLogger(outStdout, "INFO", fileWriter)
			Warning = newPlainLogger(outStderr, "WARN", fileWriter)
			Error = newPlainLogger(outStderr, "ERROR", fileWriter)
			Fatal = newPlainLogger(outStderr, "FATAL", fileWriter)
		}
		return
	}

	// Development mode
	Debug = newDevLogger(outStdout, "DEBUG", verboseMode, fileWriter)
	Info = newDevLogger(outStdout, "INFO", true, fileWriter)
	Warning = newDevLogger(outStdout, "WARN", true, fileWriter)
	Error = newDevLogger(outStdout, "ERROR", true, fileWriter)
	Fatal = newDevLogger(outStderr, "FATAL", true, fileWriter)
}

// Close closes the log file if it was opened.
// Call this function when your application shuts down to ensure logs are flushed.
func Close() error {
	if logFile != nil {
		err := logFile.Close()
		logFile = nil
		return err
	}
	return nil
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
		m[FatalLevel] = true
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
		case "FATAL":
			m[FatalLevel] = true
		}
	}
	return m
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
// If fileWriter is provided, logs are written to both console and file.
func newDevLogger(out io.Writer, level string, enabled bool, fileWriter io.Writer) *log.Logger {
	if !enabled {
		return log.New(io.Discard, "", 0)
	}
	colors := map[string]string{
		"DEBUG": "\033[36m",
		"INFO":  "\033[32m",
		"WARN":  "\033[33m",
		"ERROR": "\033[31m",
		"FATAL": "\033[35m",
	}
	reset := "\033[0m"
	levelLabel := fmt.Sprintf("%s[%s]%s", colors[level], level, reset)

	// Combine console and file output if file writer is provided
	if fileWriter != nil {
		// Write colored output to console, plain output to file
		return log.New(io.MultiWriter(out, &plainFileWriter{w: fileWriter, level: level}), levelLabel+" ", log.LstdFlags)
	}
	return log.New(out, levelLabel+" ", log.LstdFlags)
}

// newPlainLogger returns a non-colored logger for production stdout/stderr fallback.
// If fileWriter is provided, logs are written to both console and file.
func newPlainLogger(out io.Writer, level string, fileWriter io.Writer) *log.Logger {
	prefix := fmt.Sprintf("[%s] ", level)
	if fileWriter != nil {
		return log.New(io.MultiWriter(out, fileWriter), prefix, log.LstdFlags)
	}
	return log.New(out, prefix, log.LstdFlags)
}

// plainFileWriter wraps a file writer to strip ANSI color codes before writing.
type plainFileWriter struct {
	w     io.Writer
	level string
}

func (p *plainFileWriter) Write(data []byte) (int, error) {
	// Strip ANSI color codes (basic implementation)
	s := string(data)
	// Remove color codes like \033[36m and \033[0m
	var result strings.Builder
	inEscape := false
	for i := 0; i < len(s); i++ {
		if s[i] == '\033' && i+1 < len(s) && s[i+1] == '[' {
			inEscape = true
			continue
		}
		if inEscape {
			if s[i] == 'm' {
				inEscape = false
			}
			continue
		}
		result.WriteByte(s[i])
	}

	// The log.Logger already adds the level prefix, so we just need to strip colors
	// Don't add duplicate level prefix here
	return p.w.Write([]byte(result.String()))
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

// --- Formatted logging methods (fmt.Sprintf style) ---

// Debugf logs a debug message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Debugf(format string, v ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Debug.Println(msg)
}

// Infof logs an informational message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Infof(format string, v ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Info.Println(msg)
}

// Warnf logs a warning message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Warnf(format string, v ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Warning.Println(msg)
}

// Errorf logs an error message formatted with fmt.Sprintf.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Errorf(format string, v ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Error.Println(msg)
}

// Fatalf logs a fatal message formatted with fmt.Sprintf and then calls os.Exit(1).
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Fatalf(format string, v ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprintf(format, v...))
	Fatal.Println(msg)
	os.Exit(1)
}

// --- Plain logging methods (Println style) ---

// Debugln logs a debug message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Debugln(v ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Debug.Println(msg)
}

// Infoln logs an informational message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Infoln(v ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Info.Println(msg)
}

// Warnln logs a warning message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Warnln(v ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Warning.Println(msg)
}

// Errorln logs an error message by joining arguments with fmt.Sprint.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Errorln(v ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Error.Println(msg)
}

// Fatalln logs a fatal message by joining arguments with fmt.Sprint and then calls os.Exit(1).
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func Fatalln(v ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	msg := fmt.Sprintf("[%s] %s", caller, fmt.Sprint(v...))
	Fatal.Println(msg)
	os.Exit(1)
}

// --- Structured logging methods (key-value pairs) ---

// DebugKV logs a debug message with structured key-value pairs.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func DebugKV(msg string, keyvals ...any) {
	if !isLevelEnabled(DebugLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Debug.Printf("[%s] %s%s", caller, msg, fields)
}

// InfoKV logs an info message with structured key-value pairs.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func InfoKV(msg string, keyvals ...any) {
	if !isLevelEnabled(InfoLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Info.Printf("[%s] %s%s", caller, msg, fields)
}

// WarnKV logs a warning message with structured key-value pairs.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func WarnKV(msg string, keyvals ...any) {
	if !isLevelEnabled(WarnLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Warning.Printf("[%s] %s%s", caller, msg, fields)
}

// ErrorKV logs an error message with structured key-value pairs.
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func ErrorKV(msg string, keyvals ...any) {
	if !isLevelEnabled(ErrorLevel) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Error.Printf("[%s] %s%s", caller, msg, fields)
}

// FatalKV logs a fatal message with structured key-value pairs and then calls os.Exit(1).
// The caller function name and line number are automatically included.
// Thread-safe for concurrent use.
func FatalKV(msg string, keyvals ...any) {
	if !isLevelEnabled(FatalLevel) {
		os.Exit(1)
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	fields := encodeFields(keyvals...)
	Fatal.Printf("[%s] %s%s", caller, msg, fields)
	os.Exit(1)
}

// --- API logging methods (HTTP status code based) ---

// Api logs an HTTP API call with automatic level selection based on status code.
// Status codes are mapped to levels: 2xx->INFO, 4xx->WARN, 5xx->ERROR.
// Thread-safe for concurrent use.
//
// Example:
//
//	logger.Api(200, "api call successful")
//	logger.Api(404, "resource not found")
//	logger.Api(500, "internal server error")
func Api(statusCode int, msg string) {
	level := statusCodeToLevel(statusCode)
	if !isLevelEnabled(level) {
		return
	}
	logMutex.Lock()
	defer logMutex.Unlock()

	caller := getCallerInfo(2)
	logMsg := fmt.Sprintf("[%s] [%d] %s", caller, statusCode, msg)

	switch level {
	case InfoLevel:
		Info.Println(logMsg)
	case WarnLevel:
		Warning.Println(logMsg)
	case ErrorLevel:
		Error.Println(logMsg)
	}
}

// statusCodeToLevel maps HTTP status codes to log levels.
// 1xx, 2xx, 3xx -> INFO, 4xx -> WARN, 5xx -> ERROR
func statusCodeToLevel(code int) Level {
	switch {
	case code >= 500:
		return ErrorLevel
	case code >= 400:
		return WarnLevel
	case code >= 300:
		return InfoLevel // 3xx redirects are informational, not warnings
	default:
		return InfoLevel // 1xx, 2xx
	}
}
