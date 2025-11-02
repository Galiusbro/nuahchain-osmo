package logger

import (
	"fmt"
	"os"
	"strings"

	"github.com/sirupsen/logrus"
)

// Logger wraps logrus.Logger with additional configuration
type Logger struct {
	*logrus.Logger
	enabled bool
}

// Config contains logger configuration
type Config struct {
	// Enabled enables/disables logging (useful for production)
	Enabled bool

	// Level is the logging level (debug, info, warn, error)
	Level string

	// Format is the output format (json, text)
	Format string

	// Environment (dev, prod)
	Environment string
}

// New creates a new logger instance
func New(cfg Config) (*Logger, error) {
	l := logrus.New()

	if !cfg.Enabled {
		l.SetOutput(os.NewFile(0, os.DevNull)) // Discard all output
		return &Logger{Logger: l, enabled: false}, nil
	}

	// Set log level
	level, err := logrus.ParseLevel(strings.ToLower(cfg.Level))
	if err != nil {
		return nil, fmt.Errorf("invalid log level: %w", err)
	}
	l.SetLevel(level)

	// Set log format based on environment
	if cfg.Format == "json" || cfg.Environment == "prod" {
		l.SetFormatter(&logrus.JSONFormatter{
			TimestampFormat: "2006-01-02T15:04:05.000Z07:00",
		})
	} else {
		// Text formatter with colors for dev
		l.SetFormatter(&logrus.TextFormatter{
			TimestampFormat: "2006-01-02 15:04:05",
			FullTimestamp:   true,
			ForceColors:     cfg.Environment == "dev",
		})
	}

	return &Logger{
		Logger:  l,
		enabled: true,
	}, nil
}

// WithField adds a field to the logger entry
func (l *Logger) WithField(key string, value interface{}) *logrus.Entry {
	if !l.enabled {
		return logrus.NewEntry(l.Logger)
	}
	return l.Logger.WithField(key, value)
}

// WithFields adds multiple fields to the logger entry
func (l *Logger) WithFields(fields logrus.Fields) *logrus.Entry {
	if !l.enabled {
		return logrus.NewEntry(l.Logger)
	}
	return l.Logger.WithFields(fields)
}

// WithError adds an error to the logger entry
func (l *Logger) WithError(err error) *logrus.Entry {
	if !l.enabled {
		return logrus.NewEntry(l.Logger)
	}
	return l.Logger.WithError(err)
}
