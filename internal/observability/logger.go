package observability

import (
	"os"
	"strings"

	"github.com/rs/zerolog"
)

var baseLogger = zerolog.New(os.Stdout).With().Timestamp().Logger()

func ConfigureLogger(serviceName, level string) zerolog.Logger {
	logger := zerolog.New(os.Stdout).With().Timestamp().Logger()
	if strings.TrimSpace(serviceName) != "" {
		logger = logger.With().Str("service", serviceName).Logger()
	}
	if strings.TrimSpace(level) != "" {
		if parsed, err := zerolog.ParseLevel(strings.ToLower(level)); err == nil {
			logger = logger.Level(parsed)
		}
	}
	baseLogger = logger
	return logger
}

func Logger() zerolog.Logger {
	return baseLogger
}

func Logf() func(string, ...any) {
	logger := baseLogger
	return func(format string, args ...any) {
		logger.Info().Msgf(format, args...)
	}
}
