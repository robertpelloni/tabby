// Package logger provides structured logging for Tabby's Go backend.
//
// It mirrors the TypeScript Logger service with log levels and
// output to stderr (for JSON-RPC compatibility with stdin/stdout).
package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"sync"
)

// Level represents a log severity level
type Level int

const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
	LevelSilent // Suppresses all output
)

// Logger provides leveled structured logging
type Logger struct {
	mu     sync.Mutex
	level  Level
	prefix string
	out    io.Writer
	debug  *log.Logger
	info   *log.Logger
	warn   *log.Logger
	err    *log.Logger
}

// New creates a new Logger with the given prefix
func New(prefix string) *Logger {
	l := &Logger{
		level:  LevelInfo,
		prefix: prefix,
		out:    os.Stderr,
	}
	l.initLoggers()
	return l
}

func (l *Logger) initLoggers() {
	flags := log.Ldate | log.Ltime | log.Lmicroseconds
	l.debug = log.New(l.out, fmt.Sprintf("[DEBUG] [%s] ", l.prefix), flags)
	l.info = log.New(l.out, fmt.Sprintf("[INFO]  [%s] ", l.prefix), flags)
	l.warn = log.New(l.out, fmt.Sprintf("[WARN]  [%s] ", l.prefix), flags)
	l.err = log.New(l.out, fmt.Sprintf("[ERROR] [%s] ", l.prefix), flags)
}

// SetLevel changes the minimum log level
func (l *Logger) SetLevel(level Level) {
	l.mu.Lock()
	l.level = level
	l.mu.Unlock()
}

// SetOutput changes the log output writer
func (l *Logger) SetOutput(w io.Writer) {
	l.mu.Lock()
	l.out = w
	l.initLoggers()
	l.mu.Unlock()
}

// Debug logs a debug-level message
func (l *Logger) Debug(format string, args ...interface{}) {
	if l.level <= LevelDebug {
		l.debug.Printf(format, args...)
	}
}

// Info logs an info-level message
func (l *Logger) Info(format string, args ...interface{}) {
	if l.level <= LevelInfo {
		l.info.Printf(format, args...)
	}
}

// Warn logs a warning-level message
func (l *Logger) Warn(format string, args ...interface{}) {
	if l.level <= LevelWarn {
		l.warn.Printf(format, args...)
	}
}

// Error logs an error-level message
func (l *Logger) Error(format string, args ...interface{}) {
	if l.level <= LevelError {
		l.err.Printf(format, args...)
	}
}

// SubLogger creates a child logger with an extended prefix
func (l *Logger) SubLogger(suffix string) *Logger {
	return New(l.prefix + ":" + suffix)
}
