package logger

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"

	"github.com/joho/godotenv"
)

// LogLevel defines different levels of logging.
type LogLevel int

const (
	DEBUG LogLevel = iota
	INFO
	WARNING
	ERROR
)

var levelNames = []string{"DEBUG", "INFO", "WARNING", "ERROR"}

// Logger encapsulates our logging object.
type Logger struct {
	mu       sync.Mutex
	logFile  *os.File
	logLevel LogLevel
	console  bool
}

var instance *Logger
var once sync.Once

// Init initializes the logger singleton. logToConsole allows output to both console and log file.
func Init(logToConsole bool) *Logger {
	once.Do(func() {
		_ = godotenv.Load() // load any env variables
		logFilePath := "app.log"
		logLevel := getLogLevelFromEnv()

		file, err := os.OpenFile(logFilePath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
		if err != nil {
			log.Fatalf("Failed to open log file: %s", err)
		}

		instance = &Logger{
			logFile:  file,
			logLevel: logLevel,
			console:  logToConsole,
		}
	})
	return instance
}

func getLogLevelFromEnv() LogLevel {
	level := strings.ToUpper(os.Getenv("LOG_LEVEL"))
	switch level {
	case "DEBUG":
		return DEBUG
	case "INFO":
		return INFO
	case "WARNING":
		return WARNING
	case "ERROR":
		return ERROR
	default:
		return INFO
	}
}

func (l *Logger) log(level LogLevel, format string, args ...any) {
	if level < l.logLevel {
		return
	}

	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05")
	levelStr := levelNames[level]
	caller := getCallerInfo()
	message := fmt.Sprintf(format, args...)
	logMsg := fmt.Sprintf("[%s] [%s] [%s] %s\n", timestamp, levelStr, caller, message)

	if l.console {
		fmt.Print(logMsg)
	}
	_, _ = l.logFile.WriteString(logMsg)
}

func getCallerInfo() string {
	// Skip 3 frames to reach the function that called the log method.
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return "unknown"
	}
	funcName := runtime.FuncForPC(pc).Name()
	shortFile := filepath.Base(file)
	return fmt.Sprintf("%s:%d %s", shortFile, line, filepath.Base(funcName))
}

// Public API for logging.
func (l *Logger) Info(format string, args ...any) {
	l.log(INFO, format, args...)
}

func (l *Logger) Debug(format string, args ...any) {
	l.log(DEBUG, format, args...)
}

func (l *Logger) Warning(format string, args ...any) {
	l.log(WARNING, format, args...)
}

func (l *Logger) Error(format string, args ...any) {
	l.log(ERROR, format, args...)
}
