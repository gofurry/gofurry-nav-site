package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gofurry/awesome-fiber-template/v3/medium/pkg/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

// Shared logger instances.
var (
	GlobalLogger *zap.Logger
	SugarLogger  *zap.SugaredLogger
)

// Config controls logger behavior.
type Config struct {
	Level      string // debug/info/warn/error/dpanic/panic/fatal
	Mode       string // dev writes to stdout, prod writes to file
	FilePath   string // required when running in prod mode
	MaxSize    int    // single log file size in MB
	MaxBackups int    // number of rotated files to keep
	MaxAge     int    // number of days to keep old files
	Compress   bool   // compress rotated files
	ShowLine   bool   // include caller information
	EncodeJson bool   // output JSON instead of console format
	TimeFormat string // timestamp layout
	CallerSkip int    // stack depth skipped for caller display
}

func defaultConfig() Config {
	return Config{
		Level:      "info",
		Mode:       "dev",
		FilePath:   "./logs/app.log",
		MaxSize:    100,
		MaxBackups: 7,
		MaxAge:     30,
		Compress:   true,
		ShowLine:   true,
		EncodeJson: false,
		TimeFormat: common.TIME_FORMAT_DATE,
		CallerSkip: 1,
	}
}

// InitLogger initializes the global logger.
func InitLogger(cfg *Config) error {
	defaultCfg := defaultConfig()
	if cfg == nil {
		cfg = &defaultCfg
	} else {
		if cfg.Level == "" {
			cfg.Level = defaultCfg.Level
		}
		if cfg.Mode == "" {
			cfg.Mode = defaultCfg.Mode
		}
		if cfg.FilePath == "" {
			cfg.FilePath = defaultCfg.FilePath
		}
		if cfg.MaxSize == 0 {
			cfg.MaxSize = defaultCfg.MaxSize
		}
		if cfg.MaxBackups == 0 {
			cfg.MaxBackups = defaultCfg.MaxBackups
		}
		if cfg.MaxAge == 0 {
			cfg.MaxAge = defaultCfg.MaxAge
		}
		if cfg.TimeFormat == "" {
			cfg.TimeFormat = defaultCfg.TimeFormat
		}
		if cfg.CallerSkip == 0 {
			cfg.CallerSkip = defaultCfg.CallerSkip
		}
	}

	if cfg.Mode == "prod" && cfg.FilePath == "" {
		return fmt.Errorf("prod mode requires a log file path")
	}

	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return err
	}

	encoderConfig := zapcore.EncoderConfig{
		TimeKey:        "time",
		LevelKey:       "level",
		NameKey:        "logger",
		CallerKey:      "caller",
		MessageKey:     "msg",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeLevel:    zapcore.CapitalLevelEncoder,
		EncodeTime:     customTimeEncoder(cfg.TimeFormat),
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeCaller:   zapcore.ShortCallerEncoder,
	}

	var writeSyncer zapcore.WriteSyncer
	if cfg.Mode == "prod" {
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return err
		}

		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true,
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
		if !cfg.EncodeJson {
			cfg.EncodeJson = true
		}
	} else {
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	var encoder zapcore.Encoder
	if cfg.EncodeJson {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	core := zapcore.NewCore(encoder, writeSyncer, level)

	options := []zap.Option{}
	if cfg.ShowLine {
		options = append(options,
			zap.AddCaller(),
			zap.AddCallerSkip(cfg.CallerSkip),
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	}

	GlobalLogger = zap.New(core, options...)
	SugarLogger = GlobalLogger.Sugar()

	GlobalLogger.Info("logger initialized",
		String("mode", cfg.Mode),
		String("level", cfg.Level),
	)

	runtime.SetFinalizer(&GlobalLogger, func(l **zap.Logger) {
		_ = (*l).Sync()
	})

	return nil
}

func customTimeEncoder(timeFormat string) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(timeFormat))
	}
}

// Sync flushes buffered logs.
func Sync() error {
	if GlobalLogger != nil {
		return GlobalLogger.Sync()
	}
	return nil
}

// Structured fields.
func String(key, value string) zap.Field {
	return zap.String(key, value)
}

func Int(key string, value int) zap.Field {
	return zap.Int(key, value)
}

func Uint64(key string, value uint64) zap.Field {
	return zap.Uint64(key, value)
}

func Duration(key string, value time.Duration) zap.Field {
	return zap.Duration(key, value)
}

func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// Sugar logger helpers.
func Debug(args ...interface{}) {
	SugarLogger.Debug(args...)
}

func Info(args ...interface{}) {
	SugarLogger.Info(args...)
}

func Warn(args ...interface{}) {
	SugarLogger.Warn(args...)
}

func Error(args ...interface{}) {
	SugarLogger.Error(args...)
}

func Fatal(args ...interface{}) {
	SugarLogger.Fatal(args...)
}

// Formatted sugar logger helpers.
func Debugf(template string, args ...interface{}) {
	SugarLogger.Debugf(template, args...)
}

func Infof(template string, args ...interface{}) {
	SugarLogger.Infof(template, args...)
}

func Warnf(template string, args ...interface{}) {
	SugarLogger.Warnf(template, args...)
}

func Errorf(template string, args ...interface{}) {
	SugarLogger.Errorf(template, args...)
}

func Fatalf(template string, args ...interface{}) {
	SugarLogger.Fatalf(template, args...)
}

// Structured logger helpers.
func DebugWithFields(msg string, fields ...zap.Field) {
	GlobalLogger.Debug(msg, fields...)
}

func InfoWithFields(msg string, fields ...zap.Field) {
	GlobalLogger.Info(msg, fields...)
}

func WarnWithFields(msg string, fields ...zap.Field) {
	GlobalLogger.Warn(msg, fields...)
}

func ErrorWithFields(msg string, fields ...zap.Field) {
	GlobalLogger.Error(msg, fields...)
}

func FatalWithFields(msg string, fields ...zap.Field) {
	GlobalLogger.Fatal(msg, fields...)
}
