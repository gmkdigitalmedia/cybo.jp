package logger

import (
	"fmt"
	"log"
	"os"
	"time"
)

type Logger struct {
	logger *log.Logger
	level  Level
}

type Level int

const (
	DEBUG Level = iota
	INFO
	WARN
	ERROR
	FATAL
)

func New() *Logger {
	return &Logger{
		logger: log.New(os.Stdout, "", 0),
		level:  INFO,
	}
}

func (l *Logger) log(level Level, msg string, keysAndValues ...interface{}) {
	if level < l.level {
		return
	}

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	levelStr := l.levelString(level)
	
	output := fmt.Sprintf("[%s] %s: %s", timestamp, levelStr, msg)
	
	if len(keysAndValues) > 0 {
		output += " "
		for i := 0; i < len(keysAndValues); i += 2 {
			if i+1 < len(keysAndValues) {
				output += fmt.Sprintf("%v=%v ", keysAndValues[i], keysAndValues[i+1])
			}
		}
	}
	
	l.logger.Println(output)
}

func (l *Logger) levelString(level Level) string {
	switch level {
	case DEBUG:
		return "DEBUG"
	case INFO:
		return "INFO "
	case WARN:
		return "WARN "
	case ERROR:
		return "ERROR"
	case FATAL:
		return "FATAL"
	default:
		return "UNKNOWN"
	}
}

func (l *Logger) Debug(msg string, keysAndValues ...interface{}) {
	l.log(DEBUG, msg, keysAndValues...)
}

func (l *Logger) Info(msg string, keysAndValues ...interface{}) {
	l.log(INFO, msg, keysAndValues...)
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.log(WARN, msg, keysAndValues...)
}

func (l *Logger) Error(msg string, keysAndValues ...interface{}) {
	l.log(ERROR, msg, keysAndValues...)
}

func (l *Logger) Fatal(msg string, keysAndValues ...interface{}) {
	l.log(FATAL, msg, keysAndValues...)
	os.Exit(1)
}

func (l *Logger) SetLevel(level Level) {
	l.level = level
}
