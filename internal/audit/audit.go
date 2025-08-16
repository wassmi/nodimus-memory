package audit

import (
	"encoding/json"
	"io"
	"log"
	"os"
	"time"

	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger is an audit logger that writes to a rotating log file.
type Logger struct {
	*log.Logger
	writer io.WriteCloser
}

// NewLogger creates a new audit logger.
func NewLogger(filename string, maxSize, maxBackups, maxAge int, compress bool) *Logger {
	writer := &lumberjack.Logger{
		Filename:   filename,
		MaxSize:    maxSize, // megabytes
		MaxBackups: maxBackups,
		MaxAge:     maxAge,   // days
		Compress:   compress,
	}
	return &Logger{
		Logger: log.New(writer, "", 0),
		writer: writer,
	}
}

// Close closes the logger's underlying file writer.
func (l *Logger) Close() error {
	return l.writer.Close()
}


// Log records an audit event.
func (l *Logger) Log(action string, data interface{}) {
	event := struct {
		Timestamp string      `json:"timestamp"`
		Action    string      `json:"action"`
		Data      interface{} `json:"data"`
	}{
		Timestamp: time.Now().UTC().Format(time.RFC3339),
		Action:    action,
		Data:      data,
	}

	b, err := json.Marshal(event)
	if err != nil {
		log.Printf("failed to marshal audit event: %v", err)
		return
	}

	l.Println(string(b))
}

// SetFilePermissions sets the file permissions for the audit log.
func SetFilePermissions(filename string) error {
	return os.Chmod(filename, 0600)
}
