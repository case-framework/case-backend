package utils

import (
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

type LoggerConfig struct {
	LogToFile       bool   `json:"log_to_file" yaml:"log_to_file"`
	Filename        string `json:"filename" yaml:"filename"`
	MaxSize         int    `json:"max_size" yaml:"max_size"`
	MaxAge          int    `json:"max_age" yaml:"max_age"`
	MaxBackups      int    `json:"max_backups" yaml:"max_backups"`
	LogLevel        string `json:"log_level" yaml:"log_level"`
	IncludeSrc      bool   `json:"include_src" yaml:"include_src"`
	CompressOldLogs bool   `json:"compress_old_logs" yaml:"compress_old_logs"`
}

func InitLogger(
	logLevel string,
	includeSrc bool,
	logToFile bool,
	logFilename string,
	logFileMaxSize int,
	logFileMaxAge int,
	logFileMaxBackups int,
	compressOldLogs bool,
) {

	opts := &slog.HandlerOptions{
		Level:     logLevelFromString(logLevel),
		AddSource: includeSrc,
		ReplaceAttr: func(groups []string, a slog.Attr) slog.Attr {
			if a.Key == slog.SourceKey {
				source, _ := a.Value.Any().(*slog.Source)
				if source != nil {
					source.File = filepath.Base(source.File)
					source.Function = strings.Replace(source.Function, "github.com/case-framework/case-backend", "", -1)
				}
			}
			return a
		},
	}

	if logToFile && logFilename != "" {
		logTarget := &lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    logFileMaxSize,  // megabytes
			MaxAge:     logFileMaxAge,   // days
			Compress:   compressOldLogs, // compress old files
			MaxBackups: logFileMaxBackups,
		}

		w := io.MultiWriter(os.Stdout, logTarget)
		handler := slog.NewJSONHandler(w, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	} else {
		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger := slog.New(handler)
		slog.SetDefault(logger)
	}
}

func logLevelFromString(level string) slog.Level {
	switch level {
	case "debug":
		return slog.LevelDebug
	case "info":
		return slog.LevelInfo
	case "warn":
		return slog.LevelWarn
	case "error":
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func ReadConfigFromEnvAndInitLogger(
	envLogLevel string,
	envLogIncludeSrc string,
	envLogToFile string,
	envLogFilename string,
	envLogMaxSize string,
	envLogMaxAge string,
	envLogMaxBackups string,
) {
	level := os.Getenv(envLogLevel)
	includeSrc := os.Getenv(envLogIncludeSrc) == "true"
	logToFile := os.Getenv(envLogToFile) == "true"

	if !logToFile {
		InitLogger(level, includeSrc, logToFile, "", 0, 0, 0, true)
		return
	}

	logFilename := os.Getenv(envLogFilename)
	logFileMaxSize, err := strconv.Atoi(os.Getenv(envLogMaxSize))
	if err != nil {
		panic(err)
	}
	logFileMaxAge, err := strconv.Atoi(os.Getenv(envLogMaxAge))
	if err != nil {
		panic(err)
	}

	logFileMaxBackups, err := strconv.Atoi(os.Getenv(envLogMaxBackups))
	if err != nil {
		panic(err)
	}

	InitLogger(level, includeSrc, logToFile, logFilename, logFileMaxSize, logFileMaxAge, logFileMaxBackups, true)
}
