package logger

import (
	"io"
	"os"
	"raspberry_sensors/internal/config"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var log zerolog.Logger

func Get() zerolog.Logger {
	once.Do(func() {

		c := config.Get()

		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := c.Logger.Level

		var output io.Writer = zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}

		logFilePath := c.Logger.Path

		if logFilePath != "" {
			fileLogger := &lumberjack.Logger{
				Filename:   logFilePath,
				MaxSize:    1000, // 1Gb
				MaxBackups: 5,
				MaxAge:     7,
			}
			output = zerolog.MultiLevelWriter(
				output, fileLogger,
			)
		}

		log = zerolog.New(output).
			Level(zerolog.Level(logLevel)).
			With().
			Timestamp().
			Caller().
			Logger()
	})

	return log
}
