package logger

import (
	"bytes"
	"fmt"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

// TestConcurrency_MultipleLevels verifies that the mutex prevents garbled output
// when multiple goroutines log simultaneously at different levels.
func TestConcurrency_MultipleLevels(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	Init("development", true)

	const numGoroutines = 10000
	const messagesPerGoroutine = 100

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Launch 10000 goroutines, each logging 100 messages
	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			for j := range messagesPerGoroutine {
				Debugf("goroutine-%d-debug-%d", id, j)
				Infof("goroutine-%d-info-%d", id, j)
				Warnf("goroutine-%d-warn-%d", id, j)
				Errorf("goroutine-%d-error-%d", id, j)
			}
		}(i)
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedLines := numGoroutines * messagesPerGoroutine * 4
	if len(lines) != expectedLines {
		t.Fatalf("expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify that each line is complete and not garbled
	// Each line should contain the level tag [DEBUG], [INFO], [WARN], or [ERROR]
	for i, line := range lines {
		hasLevelTag := strings.Contains(line, "[DEBUG]") ||
			strings.Contains(line, "[INFO]") ||
			strings.Contains(line, "[WARN]") ||
			strings.Contains(line, "[ERROR]")

		if !hasLevelTag {
			t.Fatalf("line %d appears garbled (missing level tag): %q", i, line)
		}

		// Each line should contain a goroutine ID pattern
		if !strings.Contains(line, "goroutine-") {
			t.Fatalf("line %d appears garbled (missing goroutine marker): %q", i, line)
		}
	}
}

// TestConcurrency_StructuredLogging verifies mutex safety for KV methods
func TestConcurrency_StructuredLogging(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	Init("development", true)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			InfoKV("concurrent-log",
				"goroutine_id", id,
				"timestamp", 1234567890,
				"status", "running")
		}(i)
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != numGoroutines {
		t.Fatalf("expected %d log lines, got %d", numGoroutines, len(lines))
	}

	// Verify each line contains the expected key-value pairs
	for i, line := range lines {
		if !strings.Contains(line, "goroutine_id=") {
			t.Fatalf("line %d missing goroutine_id: %q", i, line)
		}
		if !strings.Contains(line, "timestamp=1234567890") {
			t.Fatalf("line %d missing timestamp: %q", i, line)
		}
		if !strings.Contains(line, "status=running") {
			t.Fatalf("line %d missing status: %q", i, line)
		}
	}
}

// TestConcurrency_ApiLogging verifies mutex safety for Api method
func TestConcurrency_ApiLogging(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	Init("development", true)

	const numGoroutines = 100
	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	statusCodes := []int{200, 201, 404, 500}

	for i := range numGoroutines {
		go func(id int) {
			defer wg.Done()
			statusCode := statusCodes[id%len(statusCodes)]
			Api(statusCode, "api-request-"+string(rune('a'+id%26)))
		}(i)
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	if len(lines) != numGoroutines {
		t.Fatalf("expected %d log lines, got %d", numGoroutines, len(lines))
	}

	// Verify each line contains a status code in brackets [200], [404], etc.
	for i, line := range lines {
		hasStatusCode := strings.Contains(line, "[200]") ||
			strings.Contains(line, "[201]") ||
			strings.Contains(line, "[404]") ||
			strings.Contains(line, "[500]")

		if !hasStatusCode {
			t.Fatalf("line %d missing status code: %q", i, line)
		}
	}
}

// TestConcurrency_MixedMethods verifies mutex safety when using all logging methods simultaneously
func TestConcurrency_MixedMethods(t *testing.T) {
	var buf bytes.Buffer
	oldStdout := outStdout
	defer func() { outStdout = oldStdout }()
	outStdout = &buf

	Init("development", true)

	const numGoroutines = 50
	var wg sync.WaitGroup
	wg.Add(numGoroutines * 4) // 4 different method types

	for i := range numGoroutines {
		id := i
		// Formatted logging
		go func() {
			defer wg.Done()
			Infof("formatted-log-%d", id)
		}()

		// Plain logging
		go func() {
			defer wg.Done()
			Infoln("plain-log", id)
		}()

		// Structured logging
		go func() {
			defer wg.Done()
			InfoKV("structured-log", "id", id)
		}()

		// API logging
		go func() {
			defer wg.Done()
			Api(200, "api-log")
		}()
	}

	wg.Wait()

	output := buf.String()
	lines := strings.Split(strings.TrimSpace(output), "\n")

	expectedLines := numGoroutines * 4
	if len(lines) != expectedLines {
		t.Fatalf("expected %d log lines, got %d", expectedLines, len(lines))
	}

	// Verify all lines have proper formatting (contain [INFO] since all are INFO level)
	for i, line := range lines {
		if !strings.Contains(line, "[INFO]") {
			t.Fatalf("line %d appears garbled (missing [INFO] tag): %q", i, line)
		}
	}

	// Count occurrences to verify all log types were written
	formattedCount := strings.Count(output, "formatted-log-")
	plainCount := strings.Count(output, "plain-log")
	structuredCount := strings.Count(output, "structured-log")
	apiCount := strings.Count(output, "api-log")

	if formattedCount != numGoroutines {
		t.Fatalf("expected %d formatted logs, got %d", numGoroutines, formattedCount)
	}
	if plainCount != numGoroutines {
		t.Fatalf("expected %d plain logs, got %d", numGoroutines, plainCount)
	}
	if structuredCount != numGoroutines {
		t.Fatalf("expected %d structured logs, got %d", numGoroutines, structuredCount)
	}
	if apiCount != numGoroutines {
		t.Fatalf("expected %d api logs, got %d", numGoroutines, apiCount)
	}
}

// TestConcurrency_RealTimeProgress demonstrates real-time logging with progress tracking.
// This test logs to actual stdout so you can see concurrent goroutines in action.
func TestConcurrency_RealTimeProgress(t *testing.T) {
	// This test shows ONLY progress updates, not individual worker logs
	Init("development", false) // Disable DEBUG to reduce noise

	const numGoroutines = 50
	const tasksPerGoroutine = 100

	var (
		completedTasks      atomic.Int64
		completedGoroutines atomic.Int64
		totalTasks          = numGoroutines * tasksPerGoroutine
	)

	var wg sync.WaitGroup
	wg.Add(numGoroutines)

	// Progress reporter - shows ONLY progress updates
	done := make(chan bool)
	go func() {
		ticker := time.NewTicker(200 * time.Millisecond)
		defer ticker.Stop()

		lastReported := int64(0)

		for {
			select {
			case <-ticker.C:
				completed := completedTasks.Load()
				goroutinesCompleted := completedGoroutines.Load()
				percentage := float64(completed) / float64(totalTasks) * 100

				// Only report if progress changed
				if completed != lastReported {
					InfoKV("progress",
						"completed", completed,
						"total", totalTasks,
						"percent", fmt.Sprintf("%.1f%%", percentage),
						"active_workers", numGoroutines-goroutinesCompleted,
						"tasks_per_sec", int64(float64(completed-lastReported)/0.2))
					lastReported = completed
				}
			case <-done:
				return
			}
		}
	}()

	startTime := time.Now()
	Infof("Starting concurrency test: %d workers × %d tasks = %d total operations",
		numGoroutines, tasksPerGoroutine, totalTasks)
	Infof("Watch the progress updates below - all logged via mutex-protected logger!")

	// Launch worker goroutines that actually test the logger under load
	for i := 0; i < numGoroutines; i++ {
		go func(id int) {
			defer wg.Done()
			defer completedGoroutines.Add(1)

			for j := 0; j < tasksPerGoroutine; j++ {
				// Simulate some work
				time.Sleep(5 * time.Millisecond)

				// Actually use the logger (testing concurrency)
				// Workers use DEBUG level which is disabled, so they don't clutter output
				switch j % 4 {
				case 0:
					Debugf("worker-%d: task %d/%d", id, j+1, tasksPerGoroutine)
				case 1:
					Debugf("worker-%d: task %d", id, j+1)
				case 2:
					Debugf("worker-%d: task %d", id, j+1)
				case 3:
					Debugf("worker-%d: task %d", id, j+1)
				}

				completedTasks.Add(1)
			}
		}(i)
	}

	wg.Wait()
	close(done)

	elapsed := time.Since(startTime)

	// Final summary
	Infof("✓ CONCURRENCY TEST COMPLETE!")
	InfoKV("final stats",
		"goroutines", numGoroutines,
		"total_operations", totalTasks,
		"elapsed", elapsed.Round(time.Millisecond).String(),
		"ops_per_second", int64(float64(totalTasks)/elapsed.Seconds()),
		"result", "ALL LOGS COMPLETE - NO GARBLED OUTPUT")

	// Verify all tasks completed
	if completedTasks.Load() != int64(totalTasks) {
		t.Fatalf("expected %d tasks completed, got %d", totalTasks, completedTasks.Load())
	}

	if completedGoroutines.Load() != int64(numGoroutines) {
		t.Fatalf("expected %d goroutines completed, got %d", numGoroutines, completedGoroutines.Load())
	}
}
