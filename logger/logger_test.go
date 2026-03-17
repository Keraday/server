package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestParseLogLevel(t *testing.T) {
	test := []struct {
		name     string
		input    string
		expected slog.Level
	}{
		{"debug", "debug", slog.LevelDebug},
		{"info", "info", slog.LevelInfo},
		{"warn", "warn", slog.LevelWarn},
		{"warning", "warning", slog.LevelWarn},
		{"error", "error", slog.LevelError},
		{"default", "unknown", slog.LevelInfo},
		{"empty", "", slog.LevelInfo},
	}
	for _, tt := range test {
		t.Run(tt.name, func(t *testing.T) {
			got := parseLogLevel(tt.input)
			if got != tt.expected {
				t.Errorf("parseLogLevel(%q)= %v, want: %v", tt.input, got, tt.expected)
			}
		})
	}
}

func TestNew_Prod_JSON(t *testing.T) {
	var buf bytes.Buffer
	log := New("info", true, &buf)

	log.Info("test", "key", "value")

	output := buf.String()
	if !strings.Contains(output, `"level":"INFO"`) {
		t.Errorf("expected JSON format, got: %s", output)
	}
}

func TestNew_Development_Text(t *testing.T) {
	var buf bytes.Buffer
	log := New("debug", false, &buf)

	log.Debug("test msg")

	output := buf.String()
	if !strings.Contains(output, "test msg") {
		t.Errorf("expected text format, got: %s", output)
	}
}

func TestNew_LogLevels(t *testing.T) {
	tests := []struct {
		name         string
		logLevel     string
		messageLevel string
		msg          string
		shouldOutput bool
	}{
		{"debug logger + debug msg", "debug", "debug", "test", true},
		{"debug logger + info msg", "debug", "info", "test", true},

		{"info logger + debug msg", "info", "debug", "test", false},
		{"info logger + info msg", "info", "info", "test", true},
		{"info logger + error msg", "info", "error", "test", true},

		{"error logger + info msg", "error", "info", "test", false},
		{"error logger + error msg", "error", "error", "test", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			log := New(tt.logLevel, false, &buf)

			switch tt.messageLevel {
			case "debug":
				log.Debug(tt.msg)
			case "info":
				log.Info(tt.msg)
			case "error":
				log.Error(tt.msg)
			}

			contains := strings.Contains(buf.String(), tt.msg)
			if contains != tt.shouldOutput {
				t.Errorf("log level=%q, msg level=%q: expected output=%v, got: %s",
					tt.logLevel, tt.messageLevel, tt.shouldOutput, buf.String())
			}
		})
	}
}
