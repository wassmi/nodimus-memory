package logger

import (
	"io"
	"os"
	"strings"
	"testing"

	"github.com/nodimus/nodimus/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func TestNewLoggerWithFile(t *testing.T) {
	filename := "test_logger.log"
	defer os.Remove(filename)

	cfg := config.LoggerConfig{
		File:       filename,
		MaxSize:    1,
		MaxBackups: 1,
		MaxAge:     1,
		Compress:   false,
	}

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New returned nil")
	}

	// Check if the logger's output is a lumberjack.Logger
	if _, ok := logger.Writer().(*lumberjack.Logger); !ok {
		t.Errorf("Expected logger writer to be *lumberjack.Logger, got %T", logger.Writer())
	}

	// Write something to the log and check if the file is created
	logger.Println("test log entry")
	if _, err := os.Stat(filename); os.IsNotExist(err) {
		t.Fatalf("Log file %s was not created", filename)
	}
}

func TestNewLoggerWithoutFile(t *testing.T) {
	cfg := config.LoggerConfig{
		File: "", // No file specified
	}

	// Capture stdout
	oldStdout := os.Stdout
	r, w, _ := os.Pipe()
	os.Stdout = w

	logger := New(cfg)
	if logger == nil {
		t.Fatal("New returned nil")
	}

	logger.Println("test log entry to stdout")

	w.Close()
	out, _ := io.ReadAll(r)
	os.Stdout = oldStdout // Restore stdout

	output := string(out)
	if !strings.Contains(output, "test log entry to stdout") {
		t.Errorf("Expected log entry to be written to stdout, got: %s", output)
	}
}
