package observability

import (
	"fmt"
	"log"
	"os"
	"sync"
)

// LogLevel represents logging severity
type LogLevel int

const (
	LogDebug LogLevel = iota
	LogInfo
	LogWarn
	LogError
)

// LoggerInterface defines methods for logging
type LoggerInterface interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
}

// Logger provides structured key=value logging for observability
type Logger struct {
	level  LogLevel
	logger *log.Logger
	mu     sync.Mutex
}

// Ensure Logger implements LoggerInterface
var _ LoggerInterface = (*Logger)(nil)

// NewLogger creates a logger with optional level (default: Info)
func NewLogger(level LogLevel) *Logger {
	if level < LogDebug || level > LogError {
		level = LogInfo
	}
	return &Logger{
		level:  level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

// DefaultLogger is the package-level logger used by middleware and handlers
var DefaultLogger LoggerInterface = NewLogger(LogInfo)

// internal log helper
func (l *Logger) log(level LogLevel, levelStr, msg string, keysAndValues ...interface{}) {
	if level < l.level {
		return
	}
	l.mu.Lock()
	defer l.mu.Unlock()
	buf := levelStr + " " + msg
	for i := 0; i+1 < len(keysAndValues); i += 2 {
		buf += " " + stringify(keysAndValues[i]) + "=" + stringify(keysAndValues[i+1])
	}
	l.logger.Output(3, buf)
}

// Debug logs at debug level
func (l *Logger) Debug(msg string, keysAndValues ...interface{}) { l.log(LogDebug, "DEBUG", msg, keysAndValues...) }

// Info logs at info level
func (l *Logger) Info(msg string, keysAndValues ...interface{}) { l.log(LogInfo, "INFO", msg, keysAndValues...) }

// Warn logs at warn level
func (l *Logger) Warn(msg string, keysAndValues ...interface{}) { l.log(LogWarn, "WARN", msg, keysAndValues...) }

// Error logs at error level
func (l *Logger) Error(msg string, keysAndValues ...interface{}) { l.log(LogError, "ERROR", msg, keysAndValues...) }

// helper to stringify any value
func stringify(v interface{}) string {
	switch s := v.(type) {
	case string:
		return s
	case error:
		return s.Error()
	default:
		return fmt.Sprint(v)
	}
}
