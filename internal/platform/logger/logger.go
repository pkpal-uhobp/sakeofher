package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	defaultAppLogFile   = "app.log"
	defaultErrorLogFile = "error.log"
	defaultMaxLogMB     = 50
	defaultMaxBackups   = 5
)

func New(env string) (*zap.Logger, error) {
	env = strings.ToLower(strings.TrimSpace(env))

	level := zap.NewAtomicLevelAt(zap.InfoLevel)
	if env == "local" || env == "dev" || env == "development" {
		level.SetLevel(zap.DebugLevel)
	}

	cores := []zapcore.Core{
		zapcore.NewCore(consoleEncoder(env), zapcore.Lock(os.Stdout), level),
	}

	logFolder := strings.TrimSpace(os.Getenv("LOGGER_FOLDER"))
	if logFolder != "" {
		appCore, errorCore, err := fileCores(logFolder, level)
		if err != nil {
			return nil, err
		}

		cores = append(cores, appCore, errorCore)
	}

	log := zap.New(
		zapcore.NewTee(cores...),
		zap.AddCaller(),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)

	log.Info(
		"logger initialized",
		zap.String("env", env),
		zap.String("logger_folder", logFolder),
		zap.Bool("file_logging_enabled", logFolder != ""),
		zap.String("app_log", appLogPath(logFolder)),
		zap.String("error_log", errorLogPath(logFolder)),
		zap.Int("max_log_mb", maxLogMB()),
		zap.Int("max_log_backups", maxLogBackups()),
	)

	return log, nil
}

func fileCores(logFolder string, level zap.AtomicLevel) (zapcore.Core, zapcore.Core, error) {
	if err := os.MkdirAll(logFolder, 0o755); err != nil {
		return nil, nil, fmt.Errorf("create log folder %q: %w", logFolder, err)
	}

	appWriter, err := newRotatingFileWriter(appLogPath(logFolder), maxLogMB(), maxLogBackups())
	if err != nil {
		return nil, nil, fmt.Errorf("create app log writer: %w", err)
	}

	errorWriter, err := newRotatingFileWriter(errorLogPath(logFolder), maxLogMB(), maxLogBackups())
	if err != nil {
		return nil, nil, fmt.Errorf("create error log writer: %w", err)
	}

	encoder := fileEncoder()

	appCore := zapcore.NewCore(encoder, appWriter, level)

	errorCore := zapcore.NewCore(
		encoder,
		errorWriter,
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

func maxLogMB() int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv("LOGGER_MAX_MB")))
	if err != nil || value <= 0 {
		return defaultMaxLogMB
	}

	if value > 1024 {
		return 1024
	}

	return value
}

func maxLogBackups() int {
	value, err := strconv.Atoi(strings.TrimSpace(os.Getenv("LOGGER_MAX_BACKUPS")))
	if err != nil || value < 0 {
		return defaultMaxBackups
	}

	if value > 50 {
		return 50
	}

	return value
}

type rotatingFileWriter struct {
	mu         sync.Mutex
	path       string
	maxBytes   int64
	maxBackups int
	file       *os.File
	size       int64
}

func newRotatingFileWriter(path string, maxMB int, maxBackups int) (*rotatingFileWriter, error) {
	writer := &rotatingFileWriter{
		path:       path,
		maxBytes:   int64(maxMB) * 1024 * 1024,
		maxBackups: maxBackups,
	}

	if err := writer.open(); err != nil {
		return nil, err
	}

	return writer, nil
}

func (w *rotatingFileWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.open(); err != nil {
			return 0, err
		}
	}

	if w.maxBytes > 0 && w.size+int64(len(p)) > w.maxBytes {
		if err := w.rotate(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)

	return n, err
}

func (w *rotatingFileWriter) Sync() error {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		return nil
	}

	return w.file.Sync()
}

func (w *rotatingFileWriter) open() error {
	file, err := os.OpenFile(w.path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	if err != nil {
		return err
	}

	stat, err := file.Stat()
	if err != nil {
		_ = file.Close()
		return err
	}

	w.file = file
	w.size = stat.Size()

	return nil
}

func (w *rotatingFileWriter) rotate() error {
	if w.file != nil {
		_ = w.file.Close()
		w.file = nil
	}

	if w.maxBackups <= 0 {
		_ = os.Remove(w.path)
		return w.open()
	}

	for i := w.maxBackups - 1; i >= 1; i-- {
		src := fmt.Sprintf("%s.%d", w.path, i)
		dst := fmt.Sprintf("%s.%d", w.path, i+1)

		if _, err := os.Stat(src); err == nil {
			_ = os.Rename(src, dst)
		}
	}

	if _, err := os.Stat(w.path); err == nil {
		_ = os.Rename(w.path, fmt.Sprintf("%s.1", w.path))
	}

	return w.open()
}
