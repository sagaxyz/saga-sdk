package log

import (
	"io"

	"cosmossdk.io/log"
	"github.com/pkg/errors"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
)

// This is a modified implementation of the same struct from cosmossdk.io/log.
type zeroLogWrapper struct {
	zerolog.Logger
}

func NewLogger(dst io.Writer, jsonDst io.Writer, options ...log.Option) log.Logger {
	logCfg := defaultConfig
	for _, opt := range options {
		opt(&logCfg)
	}

	var output io.Writer
	output = dst
	if !logCfg.OutputJSON {
		output = zerolog.ConsoleWriter{
			Out:        dst,
			NoColor:    !logCfg.Color,
			TimeFormat: logCfg.TimeFormat,
		}
	}

	if logCfg.Filter != nil {
		output = log.NewFilterWriter(output, logCfg.Filter)
	}

	if jsonDst != nil {
		output = zerolog.MultiLevelWriter(output, jsonDst)
	}

	logger := zerolog.New(output)
	if logCfg.StackTrace {
		zerolog.ErrorStackMarshaler = func(err error) interface{} {
			return pkgerrors.MarshalStack(errors.WithStack(err))
		}

		logger = logger.With().Stack().Logger()
	}

	if logCfg.TimeFormat != "" {
		logger = logger.With().Timestamp().Logger()
	}

	if logCfg.Level != zerolog.NoLevel {
		logger = logger.Level(logCfg.Level)
	}

	return zeroLogWrapper{logger}
}

func NewCustomLogger(logger zerolog.Logger) log.Logger {
	return zeroLogWrapper{logger}
}

func (l zeroLogWrapper) Info(msg string, keyVals ...interface{}) {
	l.Logger.Info().Fields(keyVals).Msg(msg)
}
func (l zeroLogWrapper) Error(msg string, keyVals ...interface{}) {
	l.Logger.Error().Fields(keyVals).Msg(msg)
}
func (l zeroLogWrapper) Debug(msg string, keyVals ...interface{}) {
	l.Logger.Debug().Fields(keyVals).Msg(msg)
}
func (l zeroLogWrapper) Warn(msg string, keyVals ...interface{}) {
	l.Logger.Warn().Fields(keyVals).Msg(msg)
}
func (l zeroLogWrapper) With(keyVals ...interface{}) log.Logger {
	logger := l.Logger.With().Fields(keyVals).Logger()
	return zeroLogWrapper{logger}
}
func (l zeroLogWrapper) Impl() interface{} {
	return l.Logger
}
