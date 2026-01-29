package log

import (
	"fmt"
	"io"
	"log" //nolint:depguard // logger wrapper
	"os"
	"path/filepath"
	"runtime"

	"github.com/kmdkuk/mcing/pkg/version"
)

// Level represents the severity of the log.
type Level int

const (
	// DEBUG is the debug level.
	DEBUG Level = iota
	// WARN is the warning level.
	WARN
	// ERROR is the error level.
	ERROR
	// FATAL is the fatal level.
	FATAL
)

const (
	callerSkip = 4
)

// Prefix returns the prefix string for the level.
func (level Level) Prefix() (string, error) {
	switch level {
	case DEBUG:
		return "[DEBUG] ", nil
	case WARN:
		return "[WARN] ", nil
	case ERROR:
		return "[ERROR] ", nil
	case FATAL:
		return "[FATAL] ", nil
	}
	return "", fmt.Errorf("not a valid error level %d", level)
}

var logger *Logger //nolint:gochecknoglobals // singleton

// Logger is a wrapper around [log.Logger].
type Logger struct {
	level  Level
	logger *log.Logger
}

//nolint:gochecknoinits // required by kubebuilder
func init() {
	minLevel := ERROR
	if version.Version == "DEV" {
		minLevel = DEBUG
	}
	logger = NewLogger(minLevel, os.Stdout)
}

// NewLogger creates a new Logger.
func NewLogger(l Level, w io.Writer) *Logger {
	levelPrefix, _ := l.Prefix()
	return &Logger{
		level:  l,
		logger: log.New(w, levelPrefix, log.Ldate|log.Ltime),
	}
}

// Logf logs a formatted message.
func (l *Logger) Logf(level Level, format string, v ...any) {
	l.Log(level, fmt.Sprintf(format, v...))
}

// Log logs a message.
func (l *Logger) Log(level Level, v ...any) {
	if l.IsLevelEnabled(level) {
		levelPrefix, _ := level.Prefix()
		l.logger.SetPrefix(levelPrefix)
		if level == FATAL {
			l.logger.Fatal(format(fmt.Sprint(v...)))
		}
		l.logger.Print(format(fmt.Sprint(v...)))
	}
}

// IsLevelEnabled checks if the level is enabled.
func (l *Logger) IsLevelEnabled(level Level) bool {
	return l.level <= level
}

// Debugf logs a formatted debug message.
func Debugf(format string, v ...any) {
	logger.Logf(DEBUG, format, v...)
}

// Debug logs a debug message.
func Debug(v ...any) {
	logger.Log(DEBUG, v...)
}

// Warnf logs a formatted warn message.
func Warnf(format string, v ...any) {
	logger.Logf(WARN, format, v...)
}

// Warn logs a warn message.
func Warn(v ...any) {
	logger.Log(WARN, v...)
}

// Errorf logs a formatted error message.
func Errorf(format string, v ...any) {
	logger.Logf(ERROR, format, v...)
}

// Error logs an error message.
func Error(v ...any) {
	logger.Log(ERROR, v...)
}

// Fatalf logs a formatted fatal message and exits.
func Fatalf(format string, v ...any) {
	logger.Logf(FATAL, format, v...)
}

// Fatal logs a fatal message and exits.
func Fatal(v ...any) {
	logger.Log(FATAL, v...)
}

func format(v string) string {
	_, file, line := caller()
	return fmt.Sprint(file, ":", line, " ", v)
}

func caller() (string, string, int) {
	pc, file, line, _ := runtime.Caller(callerSkip)
	f := runtime.FuncForPC(pc)
	p, _ := os.Getwd()
	path, _ := filepath.Rel(p, file)
	return f.Name(), path, line
}
