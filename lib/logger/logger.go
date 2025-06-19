package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/ariesmaulana/payroll/lib/contextutil"
	"github.com/rs/zerolog"
)

var log zerolog.Logger

// LogConfig holds configuration for the logger
type LogConfig struct {
	Debug    bool
	FilePath string
	MaxSize  int64
}

// LevelWriter implements zerolog.LevelWriter interface
type LevelWriter struct {
	file *os.File
}

// Write implements io.Writer
func (w *LevelWriter) Write(p []byte) (n int, err error) {
	return w.file.Write(p)
}

// WriteLevel implements zerolog.LevelWriter
func (w *LevelWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	return w.Write(p)
}

// Init initializes the global logger
func Init(cfg LogConfig) error {
	// Set global log level
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
	if cfg.Debug {
		zerolog.SetGlobalLevel(zerolog.DebugLevel)
	}

	// Create logs directory if it doesn't exist
	if err := os.MkdirAll(cfg.FilePath, 0755); err != nil {
		return fmt.Errorf("failed to create log directory: %w", err)
	}

	// Create or open log file
	currentTime := time.Now().Format("2006-01-02")
	logFile := filepath.Join(cfg.FilePath, fmt.Sprintf("app-%s.json", currentTime))

	file, err := os.OpenFile(logFile, os.O_CREATE|os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		return fmt.Errorf("failed to open log file: %w", err)
	}

	// Create our custom LevelWriter
	fileWriter := &LevelWriter{file: file}

	// Create multi-writer if in debug mode
	var writer zerolog.LevelWriter
	if cfg.Debug {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		writer = zerolog.MultiLevelWriter(fileWriter, consoleWriter)
	} else {
		writer = fileWriter
	}

	// Set global logger
	log = zerolog.New(writer).
		With().
		Timestamp().
		Caller().
		Logger()

	return nil
}

// Debug returns a new Event with trace information
func Debug(trace *contextutil.Trace) *zerolog.Event {
	if trace == nil {
		return log.Debug()
	}
	return log.Debug().
		Str("traceId", trace.TraceID).
		Str("method", trace.Method).
		Str("path", trace.Path).
		Str("body", trace.Body)
}

// Info returns a new Event with trace information
func Info(trace *contextutil.Trace) *zerolog.Event {
	if trace == nil {
		return log.Info()
	}
	return log.Info().
		Str("traceId", trace.TraceID).
		Str("method", trace.Method).
		Str("path", trace.Path).
		Str("body", trace.Body)
}

// Warn returns a new Event with trace information
func Warn(trace *contextutil.Trace) *zerolog.Event {
	if trace == nil {
		return log.Warn()
	}
	return log.Warn().
		Str("traceId", trace.TraceID).
		Str("method", trace.Method).
		Str("path", trace.Path).
		Str("body", trace.Body)
}

// Error returns a new Event with trace information
func Error(trace *contextutil.Trace) *zerolog.Event {
	if trace == nil {
		return log.Error()
	}
	return log.Error().
		Str("traceId", trace.TraceID).
		Str("method", trace.Method).
		Str("path", trace.Path).
		Str("body", trace.Body)
}

// Fatal returns a new Event with trace information
func Fatal(trace *contextutil.Trace) *zerolog.Event {
	if trace == nil {
		return log.Fatal()
	}
	return log.Fatal().
		Str("traceId", trace.TraceID).
		Str("method", trace.Method).
		Str("path", trace.Path).
		Str("body", trace.Body)
}
