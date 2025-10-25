package logger

import (
	"bytes"
	"github.com/coreos/go-systemd/v22/journal"
	"strings"
	"testing"
)

func TestProductionFallback_StdoutStderr(t *testing.T) {
	// Force journald disabled
	oldEnabled := journalIsEnabled
	defer func() { journalIsEnabled = oldEnabled }()
	journalIsEnabled = func() bool { return false }

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

func TestProductionJournald_Sends(t *testing.T) {
	// Force journald enabled and spy on send
	oldEnabled := journalIsEnabled
	oldSend := journalSendFunc
	defer func() { journalIsEnabled = oldEnabled; journalSendFunc = oldSend }()
	journalIsEnabled = func() bool { return true }

	type call struct {
		msg   string
		pri   journal.Priority
		ident string
	}
	var calls []call
	journalSendFunc = func(msg string, priority journal.Priority, vars map[string]string) error {
		calls = append(calls, call{msg: msg, pri: priority, ident: vars["SYSLOG_IDENTIFIER"]})
		return nil
	}

	Init("production", false)
	Infof("journal-info")
	Errorf("journal-error")

	if len(calls) < 2 {
		t.Fatalf("expected at least 2 journald sends, got %d", len(calls))
	}
	if calls[0].pri != journal.PriInfo || !strings.Contains(calls[0].msg, "journal-info") || calls[0].ident == "" {
		t.Fatalf("unexpected first call: %+v", calls[0])
	}
	if calls[1].pri != journal.PriErr || !strings.Contains(calls[1].msg, "journal-error") || calls[1].ident == "" {
		t.Fatalf("unexpected second call: %+v", calls[1])
	}
}

func TestDevelopmentVerboseTogglesDebug(t *testing.T) {
	// journald irrelevant here; just capture stdout
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
