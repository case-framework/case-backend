package utils

import (
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
)

func InitLogger(
	logLevel string,
	includeSrc bool,
	logFilename string,
	logFileMaxSize int,
	logFileMaxAge int,
	logFileMaxBackups int,
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

	if logFilename != "" {
		logTarget := &lumberjack.Logger{
			Filename:   logFilename,
			MaxSize:    logFileMaxSize, // megabytes
			MaxAge:     logFileMaxAge,  // days
			Compress:   true,           // compress old files
			MaxBackups: logFileMaxBackups,
		}
		handler := slog.NewJSONHandler(logTarget, opts)
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
