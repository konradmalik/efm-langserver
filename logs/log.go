package logs

import (
	"log"
	"os"
)

var Log Logger = &logger{createLogger(""), None}

type LogLevel int
type Logger interface {
	Logln(LogLevel, string)
	Logf(LogLevel, string, ...any)
	Printf(string, ...any)
	SetLevel(LogLevel)
}

type logger struct {
	log   *log.Logger
	level LogLevel
}

const (
	None LogLevel = iota - 1
	Error
	Warn
	Info
	Debug
)

func InitializeLogger(logfile string, level LogLevel) {
	Log = &logger{createLogger(logfile), level}
}

func (logger *logger) SetLevel(level LogLevel) {
	logger.level = level
}

func (logger logger) Logln(level LogLevel, msg string) {
	if level <= logger.level {
		logger.log.Println(msg)
	}
}

func (logger logger) Logf(level LogLevel, format string, v ...any) {
	if level <= logger.level {
		logger.log.Printf(format, v...)
	}
}

func (logger logger) Printf(format string, v ...any) {
	if logger.level >= Debug {
		logger.log.Printf(format, v...)
	}
}

func createLogger(logfile string) *log.Logger {
	if logfile == "" {
		return log.New(os.Stderr, "", log.LstdFlags)
	}

	f, err := os.OpenFile(logfile, os.O_CREATE|os.O_RDWR|os.O_APPEND, 0o660)
	if err != nil {
		log.Fatal(err)
	}
	return log.New(f, "", log.LstdFlags)
}
