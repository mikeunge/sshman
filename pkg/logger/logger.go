package logger

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/mikeunge/sshman/pkg/config"
)

type LogLevel string

const (
	DEBUG LogLevel = "DEBUG"
	INFO  LogLevel = "INFO"
	WARN  LogLevel = "WARN"
	ERROR LogLevel = "ERROR"
)

type LogEntry struct {
	Timestamp time.Time `json:"timestamp"`
	Level     LogLevel  `json:"level"`
	Message   string    `json:"message"`
	Command   string    `json:"command,omitempty"`
	SessionID string    `json:"session_id,omitempty"`
	Duration  string    `json:"duration,omitempty"`
	Error     string    `json:"error,omitempty"`
	StartTime time.Time `json:"start_time,omitempty"`
	EndTime   time.Time `json:"end_time,omitempty"`
}

type Logger struct {
	logFile   *os.File
	logMutex  sync.Mutex
	config    config.Config
}

func NewLogger(cfg config.Config) (*Logger, error) {
	logger := &Logger{
		config: cfg,
	}

	// Ensure log directory exists
	logDir := filepath.Dir(cfg.LoggingPath)
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %v", err)
	}

	// Open log file
	file, err := os.OpenFile(cfg.LoggingPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %v", err)
	}
	logger.logFile = file

	return logger, nil
}

func (l *Logger) Log(level LogLevel, message string, command string, sessionID string) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Command:   command,
		SessionID: sessionID,
	}

	l.writeLog(entry)
}

func (l *Logger) LogWithDetails(level LogLevel, message string, command string, sessionID string, duration string, startTime time.Time, endTime time.Time, err error) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     level,
		Message:   message,
		Command:   command,
		SessionID: sessionID,
		Duration:  duration,
		StartTime: startTime,
		EndTime:   endTime,
	}
	
	if err != nil {
		entry.Error = err.Error()
	}

	l.writeLog(entry)
}

func (l *Logger) LogError(message string, command string, sessionID string, err error) {
	entry := LogEntry{
		Timestamp: time.Now(),
		Level:     ERROR,
		Message:   message,
		Command:   command,
		SessionID: sessionID,
		Error:     err.Error(),
	}

	l.writeLog(entry)
}

func (l *Logger) Close() error {
	if l.logFile != nil {
		return l.logFile.Close()
	}
	return nil
}

func (l *Logger) writeLog(entry LogEntry) {
	l.logMutex.Lock()
	defer l.logMutex.Unlock()

	jsonData, err := json.Marshal(entry)
	if err != nil {
		// Fallback to console if JSON marshaling fails
		fmt.Fprintf(os.Stderr, "Failed to marshal log entry: %v\n", err)
		return
	}

	// Write JSON log entry followed by newline
	if _, err := l.logFile.Write(append(jsonData, '\n')); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to write log entry: %v\n", err)
		return
	}

	// Flush to ensure immediate write
	if err := l.logFile.Sync(); err != nil {
		fmt.Fprintf(os.Stderr, "Failed to sync log file: %v\n", err)
	}
}