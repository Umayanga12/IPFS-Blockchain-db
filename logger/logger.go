package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger defines the interface for structured logging with various levels
// and contextual logging capabilities.
type Logger interface {
	Debug(msg string, keysAndValues ...interface{})
	Info(msg string, keysAndValues ...interface{})
	Warn(msg string, keysAndValues ...interface{})
	Error(msg string, keysAndValues ...interface{})
	Fatal(msg string, keysAndValues ...interface{})
	WithFields(keysAndValues ...interface{}) Logger
	Sync() error
}

// ZapLogger implements Logger interface using Uber's Zap library.
type ZapLogger struct {
	delegate *zap.SugaredLogger
}

// Config holds logging configuration parameters.
type Config struct {
	Level      string        // Log level (debug, info, warn, error, fatal)
	Format     string        // Log format (json, console)
	BaseDir    string        // Base directory to store log files
	RotateTime time.Duration // Log rotation interval
}

// NewLogger creates a new Logger instance with specified configuration.
func NewLogger(config Config) (Logger, error) {
	if config.BaseDir == "" {
		return nil, fmt.Errorf("base directory must be specified")
	}

	// Ensure the base directory exists
	if err := os.MkdirAll(config.BaseDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create base directory: %w", err)
	}

	// Create a new log file based on the current date and time
	logFilePath := getLogFilePath(config.BaseDir)
	writer, err := getLogWriter(logFilePath, config.RotateTime)
	if err != nil {
		return nil, fmt.Errorf("failed to create log writer: %w", err)
	}

	logLevel := getZapLevel(config.Level)
	encoderConfig := getEncoderConfig(config.Format)

	core := zapcore.NewCore(
		zapcore.NewConsoleEncoder(encoderConfig),
		writer,
		logLevel,
	)

	logger := zap.New(core, zap.AddCaller(), zap.AddCallerSkip(1))
	return &ZapLogger{delegate: logger.Sugar()}, nil
}

// NewConfigFromEnv creates Config from environment variables:
// LOG_LEVEL (default: info), LOG_FORMAT (default: console), BASE_DIR (default: logs)
func NewConfigFromEnv() Config {
	level := os.Getenv("LOG_LEVEL")
	if level == "" {
		level = "info"
	}

	format := os.Getenv("LOG_FORMAT")
	if format == "" {
		format = "console"
	}

	baseDir := os.Getenv("BASE_DIR")
	if baseDir == "" {
		baseDir = "logs"
	}

	return Config{
		Level:      level,
		Format:     format,
		BaseDir:    baseDir,
		RotateTime: 10 * time.Minute,
	}
}

func getZapLevel(level string) zapcore.Level {
	switch strings.ToLower(level) {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn", "warning":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	case "fatal":
		return zapcore.FatalLevel
	default:
		return zapcore.InfoLevel
	}
}

func getEncoderConfig(format string) zapcore.EncoderConfig {
	if format == "console" {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		return cfg
	}
	return zap.NewProductionEncoderConfig()
}

func getLogFilePath(baseDir string) string {
	date := time.Now().Format("2006-01-02")
	timePart := time.Now().Format("15-04")
	dateDir := filepath.Join(baseDir, date)
	if err := os.MkdirAll(dateDir, os.ModePerm); err != nil {
		fmt.Printf("failed to create date directory: %v\n", err)
	}
	return filepath.Join(dateDir, fmt.Sprintf("log-%s.log", timePart))
}

func getLogWriter(logFilePath string, rotateTime time.Duration) (zapcore.WriteSyncer, error) {
	writer := zapcore.AddSync(&rotatingFileWriter{
		baseDir:    filepath.Dir(filepath.Dir(logFilePath)),
		rotateTime: rotateTime,
	})
	return writer, nil
}

// rotatingFileWriter handles log rotation based on time intervals.
type rotatingFileWriter struct {
	baseDir    string
	rotateTime time.Duration
	lastRotate time.Time
	file       *os.File
}

func (w *rotatingFileWriter) Write(p []byte) (n int, err error) {
	now := time.Now()
	if w.file == nil || now.Sub(w.lastRotate) >= w.rotateTime {
		if w.file != nil {
			w.file.Close()
		}
		logFilePath := getLogFilePath(w.baseDir)
		w.file, err = os.OpenFile(logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return 0, err
		}
		w.lastRotate = now
	}
	return w.file.Write(p)
}

func (w *rotatingFileWriter) Sync() error {
	if w.file != nil {
		return w.file.Sync()
	}
	return nil
}

// Debug logs a debug message with structured context.
func (l *ZapLogger) Debug(msg string, keysAndValues ...interface{}) {
	l.delegate.Debugw(msg, keysAndValues...)
}

// Info logs an info message with structured context.
func (l *ZapLogger) Info(msg string, keysAndValues ...interface{}) {
	l.delegate.Infow(msg, keysAndValues...)
}

// Warn logs a warning message with structured context.
func (l *ZapLogger) Warn(msg string, keysAndValues ...interface{}) {
	l.delegate.Warnw(msg, keysAndValues...)
}

// Error logs an error message with structured context.
func (l *ZapLogger) Error(msg string, keysAndValues ...interface{}) {
	l.delegate.Errorw(msg, keysAndValues...)
}

// Fatal logs a fatal message with structured context and exits the program.
func (l *ZapLogger) Fatal(msg string, keysAndValues ...interface{}) {
	l.delegate.Fatalw(msg, keysAndValues...)
}

// WithFields creates a new logger with additional structured context.
func (l *ZapLogger) WithFields(keysAndValues ...interface{}) Logger {
	return &ZapLogger{delegate: l.delegate.With(keysAndValues...)}
}

// Sync flushes any buffered log entries.
func (l *ZapLogger) Sync() error {
	return l.delegate.Sync()
}