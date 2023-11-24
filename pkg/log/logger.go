package log

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"

	"github.com/kmdkuk/mcing/pkg/version"
)

type Level int

const (
	DEBUG Level = iota
	WARN
	ERROR
	FATAL
)

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

var logger *Logger

type Logger struct {
	level  Level
	logger *log.Logger
}

func init() {
	minLevel := ERROR
	if version.Version == "DEV" {
		minLevel = DEBUG
	}
	logger = NewLogger(minLevel)
}

func NewLogger(l Level) *Logger {
	levelPrefix, _ := l.Prefix()
	return &Logger{
		level:  l,
		logger: log.New(os.Stdout, levelPrefix, log.Ldate|log.Ltime),
	}
}

func (l *Logger) Logf(level Level, format string, v ...interface{}) {
	l.Log(level, fmt.Sprintf(format, v...))
}

func (l *Logger) Log(level Level, v ...interface{}) {
	if l.IsLevelEnabled(level) {
		levelPrefix, _ := level.Prefix()
		l.logger.SetPrefix(levelPrefix)
		if level == FATAL {
			l.logger.Fatal(format(fmt.Sprint(v...)))
		}
		l.logger.Print(format(fmt.Sprint(v...)))
	}
}

func (l *Logger) IsLevelEnabled(level Level) bool {
	return l.level <= level
}

func Debugf(format string, v ...interface{}) {
	logger.Logf(DEBUG, format, v...)
}

func Debug(v ...interface{}) {
	logger.Log(DEBUG, v...)
}

func Warnf(format string, v ...interface{}) {
	logger.Logf(WARN, format, v...)
}

func Warn(v ...interface{}) {
	logger.Log(WARN, v...)
}

func Errorf(format string, v ...interface{}) {
	logger.Logf(ERROR, format, v...)
}

func Error(v ...interface{}) {
	logger.Log(ERROR, v...)
}

func Fatalf(format string, v ...interface{}) {
	logger.Logf(FATAL, format, v...)
}

func Fatal(v ...interface{}) {
	logger.Log(FATAL, v...)
}

func format(v string) string {
	_, file, line := caller()
	return fmt.Sprint(file, ":", line, " ", v)
}

func caller() (string, string, int) {
	pc, file, line, _ := runtime.Caller(4)
	f := runtime.FuncForPC(pc)
	p, _ := os.Getwd()
	path, _ := filepath.Rel(p, file)
	return f.Name(), path, line
}
