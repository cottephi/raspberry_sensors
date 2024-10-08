package logger

import (
	"fmt"
	"os"
	"path"
	"raspberry_sensors/internal/config"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/pkgerrors"
	"gopkg.in/natefinch/lumberjack.v2"
)

var once sync.Once

var log zerolog.Logger

type filteredWriter struct {
	w     zerolog.LevelWriter
	level zerolog.Level
}

func (w *filteredWriter) Write(p []byte) (n int, err error) {
	return w.w.Write(p)
}
func (w *filteredWriter) WriteLevel(level zerolog.Level, p []byte) (n int, err error) {
	if level >= w.level {
		return w.w.WriteLevel(level, p)
	}
	return len(p), nil
}

func Get() zerolog.Logger {
	once.Do(func() {

		c := config.Get()

		zerolog.ErrorStackMarshaler = pkgerrors.MarshalStack
		zerolog.TimeFieldFormat = time.RFC3339Nano

		logLevel := c.Logger.Level

		level, err := zerolog.ParseLevel(logLevel)
		if err != nil {
			fmt.Printf("Invalid log level: %v", err)
			os.Exit(1)
		}

		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		consoleLogger := zerolog.MultiLevelWriter(consoleWriter)
		filteredConsoleWriter := &filteredWriter{consoleLogger, level}

		logFilePath := c.Logger.Path

		if logFilePath != "" {
			fmt.Printf("Subsequent logs will be written to : %s\n", logFilePath)
			logFilePath = path.Join(logFilePath, "log.log")
			fileLogger := &lumberjack.Logger{
				Filename:   logFilePath,
				MaxSize:    1000, // 1Gb
				MaxBackups: 5,
				MaxAge:     7,
			}
			fileWriter := zerolog.MultiLevelWriter(fileLogger)
			filterdFileWriter := &filteredWriter{fileWriter, zerolog.DebugLevel}

			writer := zerolog.MultiLevelWriter(
				filteredConsoleWriter,
				filterdFileWriter,
			)

			// Initialize the logger with multi-level output
			log = zerolog.New(writer).
				With().
				Timestamp().
				Caller().
				Logger()
		} else {
			// If no file writer, just use the console logger
			fmt.Println("No logfile path provided, only logging in the console")
			log = zerolog.New(filteredConsoleWriter).
				With().
				Timestamp().
				Caller().
				Logger()
		}
	})

	return log
}
