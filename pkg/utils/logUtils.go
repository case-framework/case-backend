package utils

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"

	"gopkg.in/natefinch/lumberjack.v2"
	"gopkg.in/yaml.v2"
)

const (
	buildInfoFilename = "build-info.yaml"
	buildInfoPrefix   = "build."
)

type BuildInfoMode int

const (
	BuildInfoNever BuildInfoMode = iota
	BuildInfoOnce
	BuildInfoAlways
)

type LoggerConfig struct {
	LogToFile        bool   `json:"log_to_file" yaml:"log_to_file"`
	Filename         string `json:"filename" yaml:"filename"`
	MaxSize          int    `json:"max_size" yaml:"max_size"`
	MaxAge           int    `json:"max_age" yaml:"max_age"`
	MaxBackups       int    `json:"max_backups" yaml:"max_backups"`
	LogLevel         string `json:"log_level" yaml:"log_level"`
	IncludeSrc       bool   `json:"include_src" yaml:"include_src"`
	CompressOldLogs  bool   `json:"compress_old_logs" yaml:"compress_old_logs"`
	IncludeBuildInfo string `json:"include_build_info" yaml:"include_build_info"` // never, always, once
}

type CustomHandler struct {
	slog.Handler
	buildInfoAttrs []slog.Attr
}

func (h *CustomHandler) Handle(ctx context.Context, r slog.Record) error {
	r.AddAttrs(h.buildInfoAttrs...)
	return h.Handler.Handle(ctx, r)
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
	includeBuildInfo string,
) {

	usebuildInfo := getBuildInfoMode(includeBuildInfo)

	buildInfoAttrs := []slog.Attr{}
	if usebuildInfo != BuildInfoNever {
		buildInfoAttrs = loadBuildInfoAsSlogAttrs(buildInfoFilename, buildInfoPrefix)
	}

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

	var logger *slog.Logger
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
		logger = slog.New(handler)
	} else {
		handler := slog.NewJSONHandler(os.Stdout, opts)
		logger = slog.New(handler)
	}

	if usebuildInfo == BuildInfoAlways {
		for _, attr := range buildInfoAttrs {
			logger = logger.With(attr)
		}

	}

	slog.SetDefault(logger)

	if usebuildInfo == BuildInfoOnce {
		attrs := make([]any, len(buildInfoAttrs))
		for i, attr := range buildInfoAttrs {
			attrs[i] = attr
		}
		slog.Info("Build info", attrs...)
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

func getBuildInfoMode(includeBuildInfo string) BuildInfoMode {
	switch includeBuildInfo {
	case "never":
		return BuildInfoNever
	case "always":
		return BuildInfoAlways
	case "once":
		return BuildInfoOnce
	default:
		return BuildInfoNever
	}
}

func loadBuildInfoAsSlogAttrs(filename, prefix string) []slog.Attr {
	data, err := os.ReadFile(filename)
	if err != nil {
		panic("Error reading build info file: " + err.Error())
	}

	buildInfo := make(map[string]string)
	if err := yaml.Unmarshal(data, &buildInfo); err != nil {
		panic("Error parsing build info: " + err.Error())
	}

	attrs := make([]slog.Attr, 0, len(buildInfo))
	for k, v := range buildInfo {
		prefixedKey := fmt.Sprintf("%s%s", prefix, k)
		attrs = append(attrs, slog.String(prefixedKey, v))
	}

	return attrs
}
