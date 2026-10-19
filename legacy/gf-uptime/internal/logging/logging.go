package logging

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/gofurry/gofurry-uptime/internal/config"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

func New(cfg config.LogConfig) (*zap.Logger, error) {
	level := zapcore.InfoLevel
	if value := strings.TrimSpace(cfg.LogLevel); value != "" {
		if err := level.UnmarshalText([]byte(value)); err != nil {
			return nil, err
		}
	}
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder

	var sink zapcore.WriteSyncer
	if strings.EqualFold(strings.TrimSpace(cfg.LogMode), "prod") {
		path := strings.TrimSpace(cfg.LogPath)
		if path == "" {
			return nil, errors.New("log.log_path is required in prod mode")
		}
		if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
			return nil, err
		}
		sink = zapcore.AddSync(&lumberjack.Logger{
			Filename:   path,
			MaxSize:    positiveOr(cfg.LogMaxSize, 100),
			MaxBackups: positiveOr(cfg.LogMaxBackups, 10),
			MaxAge:     positiveOr(cfg.LogMaxAge, 7),
			Compress:   true,
			LocalTime:  true,
		})
	} else {
		sink = zapcore.AddSync(os.Stdout)
	}
	return zap.New(zapcore.NewCore(zapcore.NewJSONEncoder(encoderConfig), sink, level), zap.AddCaller()), nil
}

func positiveOr(value, fallback int) int {
	if value > 0 {
		return value
	}
	return fallback
}
