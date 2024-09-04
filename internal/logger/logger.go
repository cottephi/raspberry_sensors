package logger

import (
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"
)

// Define log levels as constants
const (
	DebugLevel = iota
	InfoLevel
	WarningLevel
	ErrorLevel
	flag = log.Ldate|log.Ltime
)

func getMaxLogSize() int64 {
	smaxsize := os.Getenv("MAX_LOG_SIZE")
	if smaxsize == "" {
		return 1 << 30  // 1GB
	}
	maxsize, err := strconv.ParseInt(smaxsize, 10, 64)
    if err != nil {
        log.Fatalf("Cannot convert MAX_LOG_SIZE value %s to int64: %s", smaxsize, err)
    }
	return maxsize
}

var MaxLogSize int64 = getMaxLogSize()

type Logger struct {
    infoLogger    *log.Logger
    warningLogger *log.Logger
    errorLogger   *log.Logger
    debugLogger   *log.Logger
		fatalLogger	  *log.Logger
    logFilePath   string
		level					int
    logFile       *os.File
    projectDir    string
}

var GlobalLogger *Logger

func InitGlobalLogger(logFilePath string, level int) error {
	var err error
	GlobalLogger, err = newLogger(logFilePath, level)
	return err
}

func newLogger(logFilePath string, level int) (*Logger, error) {

	_, file, _, ok := runtime.Caller(0)
	if !ok {
			return nil, fmt.Errorf("could not determine the project directory")
	}
	projectDir := filepath.Dir(filepath.Dir(filepath.Dir(file))) + "/"

	logger := &Logger{
		infoLogger:    log.New(os.Stdout, "INFO: ", flag),
		warningLogger: log.New(os.Stdout, "WARNING: ", flag),
		errorLogger:   log.New(os.Stderr, "ERROR: ", flag),
		debugLogger:   log.New(os.Stdout, "DEBUG: ", flag),
		fatalLogger:   log.New(os.Stderr, "FATAL: ", flag),
		logFilePath: 	 logFilePath,
		level:       	 level,
		projectDir:	   projectDir,
	}

	if err := logger.rotateLogFile(); err != nil {
			return nil, err
	}

	return logger, nil
}

// rotateLogFile checks the current log file size and rotates it if necessary
func (l *Logger) rotateLogFile() error {
	if l.logFilePath == "" {
		// Logger does not use a logfile
		return nil
	}

	fileInfo, err := os.Stat(l.logFilePath)

	if err == nil {
		// If the file exists and is over the size limit, rotate it
		if fileInfo.Size() >= MaxLogSize {
			newLogFilePath := fmt.Sprintf("%s.%s", l.logFilePath, time.Now().Format("20060102-150405"))
			if err := os.Rename(l.logFilePath, newLogFilePath); err != nil {
					return fmt.Errorf("failed to rotate log file: %v", err)
			}
		} else if l.logFile != nil {
			// File is not too big yet, continue using it
			return nil
		}
	} else if ! os.IsNotExist(err) {
		return fmt.Errorf("failed to stat log file: %v", err)
	}

	if l.logFile != nil {
		l.logFile.Close()
		l.logFile = nil
	}

	// Open a new (or existing) log file
	l.logFile, err = os.OpenFile(l.logFilePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		return fmt.Errorf("failed to open log file: %v", err)
	}

	// Create a multi-writer that writes to both stdout and the file
	out := io.MultiWriter(os.Stdout, l.logFile)
	outerr := io.MultiWriter(os.Stderr, l.logFile)

	l.debugLogger.SetOutput(l.logFile)
	l.infoLogger.SetOutput(out)
	l.warningLogger.SetOutput(out)
	l.errorLogger.SetOutput(outerr)
	l.fatalLogger.SetOutput(outerr)

	return nil
}

// logWithRotation checks if the log file needs to be rotated before logging a message
func (l *Logger) logWithRotation(f func()) {
	if err := l.rotateLogFile(); err != nil {
			fmt.Fprintf(os.Stderr, "failed to rotate log file: %v\n", err)
	}
	f()
}

func (l *Logger) Info(msg string) {
	l.logWithRotation(
		func() {
			if l.level <= InfoLevel {
				l.infoLogger.Printf("%s: %s", l.callerInfo(), msg)
			}
		},
	)
}

func (l *Logger) Infof(format string, args ...interface{}) {
	l.logWithRotation(
		func() {
			if l.level <= InfoLevel {
				l.infoLogger.Printf("%s: %s", l.callerInfo(), fmt.Sprintf(format, args...))
			}
		},
	)
}

func (l *Logger) Warning(msg string) {
	l.logWithRotation(
		func() {
			if l.level <= WarningLevel {
					l.warningLogger.Printf("%s: %s", l.callerInfo(), msg)
			}
		},
	)
}

func (l *Logger) Warningf(format string, args ...interface{}) {
	l.logWithRotation(
		func() {
			if l.level <= WarningLevel {
				l.warningLogger.Printf("%s: %s", l.callerInfo(), fmt.Sprintf(format, args...))
			}
		},
	)
}

func (l *Logger) Error(msg interface{}) {
	l.logWithRotation(
		func() {
			if l.level <= ErrorLevel {	
				l.handleError(msg, l.errorLogger)
			}
		},
	)
}

func (l *Logger) Errorf(format string, args ...interface{}) {
	l.logWithRotation(
		func() {
			if l.level <= ErrorLevel {
				l.errorLogger.Printf("%s: %s", l.callerInfo(), fmt.Sprintf(format, args...))
			}
		},
	)
}

func (l *Logger) Debug(msg string) {
	l.logWithRotation(
		func() {
			if l.level <= DebugLevel {
					l.debugLogger.Printf("%s: %s", l.callerInfo(), msg)
			}
		},
	)
}

func (l *Logger) Debugf(format string, args ...interface{}) {
	l.logWithRotation(
		func() {
			if l.level <= DebugLevel {
				l.debugLogger.Printf("%s: %s", l.callerInfo(), fmt.Sprintf(format, args...))
			}
		},
	)
}

func (l *Logger) Fatal(msg interface{}) {
	l.logWithRotation(
		func() {
			l.handleError(msg, l.fatalLogger)
			os.Exit(1)
		},
	)
}

func (l *Logger) Fatalf(format string, args ...interface{}) {
	l.logWithRotation(
		func() {
			l.fatalLogger.Printf("%s: %s", l.callerInfo(), fmt.Sprintf(format, args...))
			os.Exit(1)
		},
	)
}

// Close closes the log file if necessary
func (l *Logger) Close() {
    if l.logFile != nil {
        l.logFile.Close()
    }
}


func (l *Logger) handleError(msg interface{}, logger *log.Logger) {
	switch msg := msg.(type) {
	case string:
		logger.Printf("%s: %s", l.callerInfo(), msg)
	case error:
		logger.Println(msg.Error())
	default:
		logger.Printf("%s: %s", l.callerInfo(), msg)
	}
}

// callerInfo retrieves the file and line number of the caller
func (l *Logger) callerInfo() string {

	file := "logger.go"
	line := 0
	ok := true
	i := 0

	for strings.HasSuffix(file, "logger.go") {
		_, file, line, ok = runtime.Caller(i)
		if !ok {
				return "unknown:0"
		}
		i++
	}

	file = strings.TrimPrefix(file, l.projectDir)

	return fmt.Sprintf("%s:%d", file, line)
}