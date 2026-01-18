package observability

import (
	"os"
	"time"

	"github.com/rs/zerolog"
)

func NewLogger(debug bool) (*zerolog.Logger, error) {
	if debug {
		return NewDevelopmentLogger()
	}
	return NewProductionLogger()
}

func NewDevelopmentLogger() (*zerolog.Logger, error) {
	writer := zerolog.ConsoleWriter{
		Out:        os.Stdout,
		TimeFormat: time.RFC3339,
	}

	logger := zerolog.New(writer).
		With().
		Timestamp().
		Logger().
		Level(zerolog.DebugLevel)

	return &logger, nil
}

func NewProductionLogger() (*zerolog.Logger, error) {
	logger := zerolog.New(os.Stdout).
		With().
		Timestamp().
		Logger().
		Level(zerolog.InfoLevel)

	return &logger, nil
}
