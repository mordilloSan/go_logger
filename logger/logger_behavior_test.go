package logger

import (
	"bytes"
	"strings"
	"testing"
)

func TestProductionFallback_StdoutStderr(t *testing.T) {
	// Capture stdout/stderr
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout, oldStderr := outStdout, outStderr
	defer func() { outStdout, outStderr = oldStdout, oldStderr }()
	outStdout = &stdoutBuf
	outStderr = &stderrBuf

	Init("production", false)

	Infof("hello")
	Debugf("dbg")
	Warnf("careful")
	Errorf("boom")

	if got := stdoutBuf.String(); !strings.Contains(got, "hello") || !strings.Contains(got, "dbg") {
		t.Fatalf("stdout missing expected logs, got: %q", got)
	}
	if got := stderrBuf.String(); !strings.Contains(got, "careful") || !strings.Contains(got, "boom") {
		t.Fatalf("stderr missing expected logs, got: %q", got)
	}
}

func TestProductionPlainOutput(t *testing.T) {
	var stdoutBuf, stderrBuf bytes.Buffer
	oldStdout, oldStderr := outStdout, outStderr
	defer func() { outStdout, outStderr = oldStdout, oldStderr }()
	outStdout = &stdoutBuf
	outStderr = &stderrBuf

	Init("production", false)
	Infof("prod-info")
	Errorf("prod-error")

	if got := stdoutBuf.String(); !strings.Contains(got, "prod-info") {
		t.Fatalf("stdout missing expected logs, got: %q", got)
	}
	if got := stderrBuf.String(); !strings.Contains(got, "prod-error") {
		t.Fatalf("stderr missing expected logs, got: %q", got)
	}
	if strings.Contains(stdoutBuf.String(), "\033[") || strings.Contains(stderrBuf.String(), "\033[") {
		t.Fatalf("production output should be plain (no ANSI codes), got stdout=%q stderr=%q", stdoutBuf.String(), stderrBuf.String())
	}
}

func TestDevelopmentVerboseTogglesDebug(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()

	outStdout = &buf
	Init("development", false)
	Debugf("dev-false")
	if got := buf.String(); strings.Contains(got, "dev-false") {
		t.Fatalf("debug should be disabled in development when verbose=false, got: %q", got)
	}

	buf.Reset()
	outStdout = &buf
	Init("development", true)
	Debugf("dev-true")
	if got := buf.String(); !strings.Contains(got, "dev-true") {
		t.Fatalf("debug should be enabled in development when verbose=true, got: %q", got)
	}
}
