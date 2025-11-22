package logger

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestFileLogging_Development(t *testing.T) {
	// Create a temporary log file
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "test.log")

	// Initialize logger with file logging
	InitWithFile("development", true, logPath)
	defer Close()

	// Log some messages
	Infof("test info message")
	Warnln("test warning")
	ErrorKV("test error", "key", "value")

	// Read the log file
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Verify all messages are in the file
	if !strings.Contains(log, "test info message") {
		t.Errorf("log file should contain info message, got: %q", log)
	}
	if !strings.Contains(log, "test warning") {
		t.Errorf("log file should contain warning, got: %q", log)
	}
	if !strings.Contains(log, "test error") {
		t.Errorf("log file should contain error message, got: %q", log)
	}
	if !strings.Contains(log, "key=value") {
		t.Errorf("log file should contain structured fields, got: %q", log)
	}

	// Verify level prefixes are present
	if !strings.Contains(log, "[INFO]") {
		t.Errorf("log file should contain [INFO] prefix, got: %q", log)
	}
	if !strings.Contains(log, "[WARN]") {
		t.Errorf("log file should contain [WARN] prefix, got: %q", log)
	}
	if !strings.Contains(log, "[ERROR]") {
		t.Errorf("log file should contain [ERROR] prefix, got: %q", log)
	}

	// Verify no ANSI color codes in file
	if strings.Contains(log, "\033[") {
		t.Errorf("log file should not contain ANSI color codes, got: %q", log)
	}
}

func TestFileLogging_Production(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "prod.log")

	InitWithFile("production", false, logPath)
	defer Close()

	Infof("production info")
	Errorf("production error")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	if !strings.Contains(log, "production info") {
		t.Errorf("log should contain production info, got: %q", log)
	}
	if !strings.Contains(log, "production error") {
		t.Errorf("log should contain production error, got: %q", log)
	}
}

func TestFileLogging_Append(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "append.log")

	// First initialization
	InitWithFile("development", true, logPath)
	Infof("first message")
	Close()

	// Second initialization (should append, not overwrite)
	InitWithFile("development", true, logPath)
	Infof("second message")
	Close()

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Both messages should be present
	if !strings.Contains(log, "first message") {
		t.Errorf("log should contain first message, got: %q", log)
	}
	if !strings.Contains(log, "second message") {
		t.Errorf("log should contain second message, got: %q", log)
	}
}

func TestFileLogging_InvalidPath(t *testing.T) {
	// Try to log to an invalid path (should not crash, just continue without file logging)
	invalidPath := "/nonexistent/directory/test.log"

	// This should not crash - it will print an error to stderr but continue
	InitWithFile("development", true, invalidPath)
	defer Close()

	// Logger should still work (just won't write to file)
	Infof("test message")

	// The logFile should be nil since the file couldn't be opened
	if logFile != nil {
		t.Errorf("logFile should be nil when path is invalid, got: %v", logFile)
	}
}

func TestFileLogging_Close(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "close.log")

	InitWithFile("development", true, logPath)
	Infof("before close")

	// Close should succeed
	err := Close()
	if err != nil {
		t.Errorf("Close() should not return error, got: %v", err)
	}

	// Second close should be safe (no-op)
	err = Close()
	if err != nil {
		t.Errorf("second Close() should not return error, got: %v", err)
	}

	// File should contain the message
	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	if !strings.Contains(string(content), "before close") {
		t.Errorf("log should contain message, got: %q", string(content))
	}
}

func TestFileLogging_NoFile(t *testing.T) {
	// Init without file (empty path) should work normally
	InitWithFile("development", true, "")
	defer Close()

	// Should not crash
	Infof("test message")

	// Close should be safe even with no file
	err := Close()
	if err != nil {
		t.Errorf("Close() with no file should not error, got: %v", err)
	}
}

func TestFileLogging_AllLevels(t *testing.T) {
	tmpDir := t.TempDir()
	logPath := filepath.Join(tmpDir, "levels.log")

	InitWithFile("development", true, logPath)
	defer Close()

	// Test all logging methods
	Debugf("debug %s", "formatted")
	Debugln("debug plain")
	DebugKV("debug structured", "key", "value")

	Infof("info %s", "formatted")
	Infoln("info plain")
	InfoKV("info structured", "key", "value")

	Warnf("warn %s", "formatted")
	Warnln("warn plain")
	WarnKV("warn structured", "key", "value")

	Errorf("error %s", "formatted")
	Errorln("error plain")
	ErrorKV("error structured", "key", "value")

	content, err := os.ReadFile(logPath)
	if err != nil {
		t.Fatalf("failed to read log file: %v", err)
	}

	log := string(content)

	// Verify all levels are logged
	expectedStrings := []string{
		"debug formatted", "debug plain", "debug structured",
		"info formatted", "info plain", "info structured",
		"warn formatted", "warn plain", "warn structured",
		"error formatted", "error plain", "error structured",
		"[DEBUG]", "[INFO]", "[WARN]", "[ERROR]",
	}

	for _, expected := range expectedStrings {
		if !strings.Contains(log, expected) {
			t.Errorf("log should contain %q, got: %q", expected, log)
		}
	}
}

func TestInit_BackwardCompatible(t *testing.T) {
	// Test that the old Init function still works (no file logging)
	Init("development", true)
	defer Close()

	// Should not crash
	Infof("test message")

	// Close should be safe
	err := Close()
	if err != nil {
		t.Errorf("Close() should not error, got: %v", err)
	}
}
