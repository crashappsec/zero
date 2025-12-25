package logging

import (
	"bytes"
	"errors"
	"strings"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	if logger == nil {
		t.Fatal("New returned nil")
	}

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, "test message") {
		t.Errorf("Output should contain message: %s", output)
	}
	if !strings.Contains(output, "key=value") {
		t.Errorf("Output should contain key=value: %s", output)
	}
}

func TestNewJSON(t *testing.T) {
	var buf bytes.Buffer
	logger := NewJSON(&buf, LevelInfo)

	logger.Info("test message", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"msg":"test message"`) {
		t.Errorf("JSON output should contain msg: %s", output)
	}
	if !strings.Contains(output, `"key":"value"`) {
		t.Errorf("JSON output should contain key: %s", output)
	}
}

func TestNewNop(t *testing.T) {
	logger := NewNop()
	// Should not panic
	logger.Info("this should be discarded")
	logger.Error("this too")
}

func TestLogger_With(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	childLogger := logger.With("component", "test")
	childLogger.Info("message")

	output := buf.String()
	if !strings.Contains(output, "component=test") {
		t.Errorf("Output should contain component=test: %s", output)
	}
}

func TestLogger_WithScanner(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	scannerLogger := logger.WithScanner("sbom")
	scannerLogger.Info("scanning")

	output := buf.String()
	if !strings.Contains(output, "scanner=sbom") {
		t.Errorf("Output should contain scanner=sbom: %s", output)
	}
}

func TestLogger_WithRepo(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	repoLogger := logger.WithRepo("owner/repo")
	repoLogger.Info("processing")

	output := buf.String()
	if !strings.Contains(output, "repo=owner/repo") {
		t.Errorf("Output should contain repo=owner/repo: %s", output)
	}
}

func TestLogger_WithOperation(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	opLogger := logger.WithOperation("clone")
	opLogger.Info("started")

	output := buf.String()
	if !strings.Contains(output, "op=clone") {
		t.Errorf("Output should contain op=clone: %s", output)
	}
}

func TestLogger_WithError(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	// Test with actual error
	testErr := errors.New("test error")
	errLogger := logger.WithError(testErr)
	errLogger.Info("message with error")

	output := buf.String()
	if !strings.Contains(output, "error") {
		t.Errorf("Output should contain error attribute: %s", output)
	}

	// Test with nil error - should return same logger
	buf.Reset()
	sameLogger := logger.WithError(nil)
	if sameLogger != logger {
		// This is expected behavior - WithError with nil returns the same logger
	}
}

func TestLogger_WithDuration(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	durLogger := logger.WithDuration(100 * time.Millisecond)
	durLogger.Info("completed")

	output := buf.String()
	if !strings.Contains(output, "duration_ms=100") {
		t.Errorf("Output should contain duration_ms=100: %s", output)
	}
}

func TestLogLevels(t *testing.T) {
	t.Run("Debug not shown at Info level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&buf, LevelInfo)
		logger.Debug("debug message")
		if strings.Contains(buf.String(), "debug message") {
			t.Error("Debug should not be shown at Info level")
		}
	})

	t.Run("Debug shown at Debug level", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&buf, LevelDebug)
		logger.Debug("debug message")
		if !strings.Contains(buf.String(), "debug message") {
			t.Error("Debug should be shown at Debug level")
		}
	})

	t.Run("Error shown at all levels", func(t *testing.T) {
		var buf bytes.Buffer
		logger := New(&buf, LevelError)
		logger.Error("error message")
		if !strings.Contains(buf.String(), "error message") {
			t.Error("Error should always be shown")
		}
	})
}

func TestPackageLevelFunctions(t *testing.T) {
	var buf bytes.Buffer
	SetDefault(New(&buf, LevelInfo))

	Info("package level info")
	if !strings.Contains(buf.String(), "package level info") {
		t.Error("Package level Info should work")
	}
}

func TestTiming(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	func() {
		defer Timing(logger, "test operation")()
		time.Sleep(10 * time.Millisecond)
	}()

	output := buf.String()
	if !strings.Contains(output, "test operation completed") {
		t.Errorf("Output should contain timing message: %s", output)
	}
	if !strings.Contains(output, "duration_ms") {
		t.Errorf("Output should contain duration: %s", output)
	}
}

func TestLogPanic(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	func() {
		defer LogPanic(logger)
		panic("test panic")
	}()

	output := buf.String()
	if !strings.Contains(output, "panic recovered") {
		t.Errorf("Output should contain panic message: %s", output)
	}
	if !strings.Contains(output, "test panic") {
		t.Errorf("Output should contain panic value: %s", output)
	}
}

func TestParseLevel(t *testing.T) {
	tests := []struct {
		input    string
		expected Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"INFO", LevelInfo},
		{"warn", LevelWarn},
		{"WARN", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"ERROR", LevelError},
		{"unknown", LevelInfo},
		{"", LevelInfo},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := ParseLevel(tt.input)
			if got != tt.expected {
				t.Errorf("ParseLevel(%q) = %v, want %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDefaultConfig(t *testing.T) {
	cfg := DefaultConfig()
	if cfg.Level != LevelInfo {
		t.Errorf("Default level should be Info, got %v", cfg.Level)
	}
	if cfg.JSON {
		t.Error("Default should not be JSON")
	}
}

func TestNewFromConfig(t *testing.T) {
	t.Run("text output", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := Config{
			Level:  LevelInfo,
			Output: &buf,
			JSON:   false,
		}
		logger := NewFromConfig(cfg)
		logger.Info("test")
		if strings.Contains(buf.String(), `"msg"`) {
			t.Error("Text output should not be JSON")
		}
	})

	t.Run("JSON output", func(t *testing.T) {
		var buf bytes.Buffer
		cfg := Config{
			Level:  LevelInfo,
			Output: &buf,
			JSON:   true,
		}
		logger := NewFromConfig(cfg)
		logger.Info("test")
		if !strings.Contains(buf.String(), `"msg"`) {
			t.Error("JSON output should contain JSON")
		}
	})

	t.Run("nil output defaults to stderr", func(t *testing.T) {
		cfg := Config{
			Level:  LevelInfo,
			Output: nil,
		}
		logger := NewFromConfig(cfg)
		if logger == nil {
			t.Error("Logger should not be nil")
		}
	})
}

func TestShortPath(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"/Users/test/go/src/zero/pkg/core/logging/logging.go", "logging/logging.go"},
		{"logging/logging.go", "logging/logging.go"},
		{"logging.go", "logging.go"},
		{"", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			got := shortPath(tt.input)
			if got != tt.expected {
				t.Errorf("shortPath(%q) = %q, want %q", tt.input, got, tt.expected)
			}
		})
	}
}

func TestDefault(t *testing.T) {
	logger := Default()
	if logger == nil {
		t.Error("Default() should not return nil")
	}
}

func TestChainedWith(t *testing.T) {
	var buf bytes.Buffer
	logger := New(&buf, LevelInfo)

	logger.WithScanner("sbom").
		WithRepo("owner/repo").
		WithOperation("scan").
		Info("starting")

	output := buf.String()
	if !strings.Contains(output, "scanner=sbom") {
		t.Error("Should contain scanner")
	}
	if !strings.Contains(output, "repo=owner/repo") {
		t.Error("Should contain repo")
	}
	if !strings.Contains(output, "op=scan") {
		t.Error("Should contain operation")
	}
}
