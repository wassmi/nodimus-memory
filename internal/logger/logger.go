package logger

import (
	"io"
	"log"
	"os"

	"github.com/nodimus/nodimus/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

// New creates a new logger based on the provided configuration.
func New(cfg config.LoggerConfig) *log.Logger {
	var output io.Writer
	if cfg.File != "" {
		output = &lumberjack.Logger{
			Filename:   cfg.File,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
		}
	} else {
		output = os.Stdout
	}

	return log.New(output, "", log.LstdFlags)
}
