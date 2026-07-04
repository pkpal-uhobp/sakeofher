package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultAppLogFile   = "app.log"
	defaultErrorLogFile = "error.log"
)

func New(env string) (*zap.Logger, error) {
	env = strings.ToLower(strings.TrimSpace(env))

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if env == "local" || env == "dev" || env == "development" {
		level.SetLevel(zap.DebugLevel)
	}

	cores := make([]zapcore.Core, 0, 3)

	// Console output stays enabled so `make run-api` and `make run-worker`
	// keep showing logs in the terminal.
	cores = append(cores, zapcore.NewCore(
		consoleEncoder(env),
		zapcore.Lock(os.Stdout),
		level,
	))

	logFolder := strings.TrimSpace(os.Getenv("LOGGER_FOLDER"))
	if logFolder != "" {
		appCore, errorCore, err := fileCores(logFolder, level)
		if err != nil {
			return nil, err
		}

		cores = append(cores, appCore, errorCore)
	}

	options := []zap.Option{
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	}

	log := zap.New(zapcore.NewTee(cores...), options...)

	log.Info(
		"logger initialized",
		zap.String("env", env),
		zap.String("logger_folder", logFolder),
		zap.Bool("file_logging_enabled", logFolder != ""),
		zap.String("app_log", appLogPath(logFolder)),
		zap.String("error_log", errorLogPath(logFolder)),
	)

	return log, nil
}

func fileCores(logFolder string, level zap.AtomicLevel) (zapcore.Core, zapcore.Core, error) {
	if err := os.MkdirAll(logFolder, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create log folder %q: %w", logFolder, err)
	}

	appFile, err := os.OpenFile(appLogPath(logFolder), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return nil, nil, fmt.Errorf("open app log file: %w", err)
	}

	errorFile, err := os.OpenFile(errorLogPath(logFolder), os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		_ = appFile.Close()
		return nil, nil, fmt.Errorf("open error log file: %w", err)
	}

	encoder := fileEncoder()

	appCore := zapcore.NewCore(
		encoder,
		zapcore.Lock(appFile),
		level,
	)

	errorCore := zapcore.NewCore(
		encoder,
		zapcore.Lock(errorFile),
		zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
			return lvl >= zapcore.ErrorLevel
		}),
	)

	return appCore, errorCore, nil
}

func consoleEncoder(env string) zapcore.Encoder {
	if env == "local" || env == "dev" || env == "development" {
		cfg := zap.NewDevelopmentEncoderConfig()
		cfg.EncodeLevel = zapcore.CapitalColorLevelEncoder
		cfg.EncodeTime = zapcore.ISO8601TimeEncoder
		cfg.EncodeDuration = zapcore.StringDurationEncoder

		return zapcore.NewConsoleEncoder(cfg)
	}

	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder

	return zapcore.NewJSONEncoder(cfg)
}

func fileEncoder() zapcore.Encoder {
	cfg := zap.NewProductionEncoderConfig()
	cfg.EncodeTime = zapcore.ISO8601TimeEncoder
	cfg.EncodeDuration = zapcore.StringDurationEncoder
	cfg.TimeKey = "ts"
	cfg.LevelKey = "level"
	cfg.NameKey = "logger"
	cfg.CallerKey = "caller"
	cfg.MessageKey = "msg"
	cfg.StacktraceKey = "stacktrace"

	return zapcore.NewJSONEncoder(cfg)
}

func appLogPath(logFolder string) string {
	if strings.TrimSpace(logFolder) == "" {
		return ""
	}

	return filepath.Join(logFolder, defaultAppLogFile)
}

func errorLogPath(logFolder string) string {
	if strings.TrimSpace(logFolder) == "" {
		return ""
	}

	return filepath.Join(logFolder, defaultErrorLogFile)
}
