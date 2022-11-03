package logger

import (
	"io"

	"github.com/rs/zerolog"
)

type Logger struct {
	logger zerolog.Logger
}

func NewLogger(w io.Writer) *Logger {
	r := &Logger{}
	r.logger = zerolog.New(w).With().Timestamp().Logger()

	return r
}

func (r *Logger) Fatal(message string) {
	r.logger.Fatal().Interface("message", message).Send()
}

func (r *Logger) Panic(message string) {
	r.logger.Panic().Interface("message", message).Send()
}

func (r *Logger) Warning(message string) {
	r.logger.Warn().Interface("message", message).Send()
}

func (r *Logger) Info(message string) {
	r.logger.Info().Interface("message", message).Send()
}
