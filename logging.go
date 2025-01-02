package iapetus

import (
	"log"
	"os"
)

// LogLevel represents the severity of a log message
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	ERROR
)

type Logger interface {
	Debug(format string, args ...interface{})
	Info(format string, args ...interface{})
	Error(format string, args ...interface{})
	SetLevel(level LogLevel)
}

// DefaultLogger implements Logger using the standard log package
type DefaultLogger struct {
	level  LogLevel
	logger *log.Logger
}

// NewDefaultLogger creates a new DefaultLogger with INFO as default level
func NewDefaultLogger(level *LogLevel) *DefaultLogger {
	return &DefaultLogger{
		level:  *level,
		logger: log.New(os.Stdout, "", log.LstdFlags),
	}
}

func (l *DefaultLogger) Debug(format string, args ...interface{}) {
	if l.level <= DEBUG {
		l.logger.Printf("[DEBUG] "+format, args...)
	}
}

func (l *DefaultLogger) Info(format string, args ...interface{}) {
	if l.level == INFO {
		l.logger.Printf("[INFO] "+format, args...)
	}
}

func (l *DefaultLogger) Error(format string, args ...interface{}) {
	if l.level == ERROR {
		l.logger.Printf("[ERROR] "+format, args...)
	}
}

func (l *DefaultLogger) SetLevel(level LogLevel) {
	l.level = level
}
