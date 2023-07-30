package logger

import (
	"io"
	"os"
	"sync"
	"time"

	"github.com/rs/zerolog"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Logger provides functionality for logging.
type Logger struct {
	*zerolog.Logger
}

func newFileWriter(filename string) io.Writer {
	return &lumberjack.Logger{
		Filename: filename,
	}
}

var (
	logger Logger
	once   sync.Once
)

// New returns a new instance of logger.
func New(logLevel, logFilename string) *Logger {
	once.Do(func() {
		// By default create console writer
		writers := []io.Writer{zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp}}

		if logFilename != "" {
			writers = append(writers, newFileWriter(logFilename))
		}

		if logLevel != "" {
			level, err := zerolog.ParseLevel(logLevel)
			if err != nil {
				panic(err)
			}

			zerolog.SetGlobalLevel(level)
		}

		multiWriters := io.MultiWriter(writers...)

		zeroLogger := zerolog.New(multiWriters).With().Caller().Timestamp().Logger()

		logger = Logger{&zeroLogger}
	})

	return &logger
}
