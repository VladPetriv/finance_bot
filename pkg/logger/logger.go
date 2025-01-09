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

// LoggergerOptions represents options for logger.
type LoggergerOptions struct {
	LogLevel        string
	LogFile         string
	PrettyLogOutput bool
}

// New returns a new instance of logger.
func New(opts LoggergerOptions) *Logger {
	once.Do(func() {
		// By default create console writer
		writers := []io.Writer{os.Stdout}

		if opts.PrettyLogOutput {
			writers[0] = zerolog.ConsoleWriter{Out: os.Stdout, TimeFormat: time.Stamp}

		}

		if opts.LogFile != "" {
			writers = append(writers, newFileWriter(opts.LogFile))
		}

		if opts.LogLevel != "" {
			level, err := zerolog.ParseLevel(opts.LogLevel)
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
