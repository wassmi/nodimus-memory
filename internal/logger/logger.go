package logger

import (
	"log"
	"os"
	"path/filepath"

	"github.com/wassmi/nodimus-memory/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a new logger. If the log file path in the config is not absolute,
// it is resolved relative to the provided dataDir.
func New(cfg config.LoggerConfig, dataDir string) *log.Logger {
	logFilePath := cfg.File
	if logFilePath != "" && !filepath.IsAbs(logFilePath) {
		logFilePath = filepath.Join(dataDir, logFilePath)
	}

	// Ensure the directory for the log file exists.
	if logFilePath != "" {
		if err := os.MkdirAll(filepath.Dir(logFilePath), 0755); err != nil {
			log.Fatalf("failed to create log directory: %v", err)
		}
	}

	writer := &lumberjack.Logger{
		Filename:   logFilePath,
		MaxSize:    cfg.MaxSize,
		MaxBackups: cfg.MaxBackups,
		MaxAge:     cfg.MaxAge,
		Compress:   cfg.Compress,
	}

	// If no file is specified, log to stdout.
	if cfg.File == "" {
		return log.New(os.Stdout, "", log.LstdFlags)
	}

	return log.New(writer, "", log.LstdFlags)
}