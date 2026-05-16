package log

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"time"

	"github.com/gofurry/gofurry-game-backend/common"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

/*
 * @Desc: 日志 Zap封装
 * @author: 福狼
 * @version: v1.0.1
 */

// 全局日志实例
var (
	GlobalLogger *zap.Logger
	SugarLogger  *zap.SugaredLogger
)

// Config 日志配置结构体
type Config struct {
	Level      string // 日志级别 debug/info/warn/error/dpanic/panic/fatal
	Mode       string // 运行模式 dev(控制台)/prod(文件)
	FilePath   string // 日志文件路径(prod必填)
	MaxSize    int    // 单个日志文件大小(MB)
	MaxBackups int    // 最大备份数
	MaxAge     int    // 最大保留天数
	Compress   bool   // 是否压缩备份
	ShowLine   bool   // 是否显示代码行号
	EncodeJson bool   // 是否 JSON 格式输出
	TimeFormat string // 时间格式
	CallerSkip int    // 调用栈跳过层数
}

// defaultConfig 获取默认日志配置
func defaultConfig() Config {
	return Config{
		Level:      "info",
		Mode:       "dev",
		FilePath:   "./logs/gf-steam-sdk.log",
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

// InitLogger 初始化日志
func InitLogger(cfg *Config) error {
	// 合并默认配置
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

	// 验证生产模式配置
	if cfg.Mode == "prod" && cfg.FilePath == "" {
		return fmt.Errorf("生产模式必须配置文件路径")
	}

	// 设置日志级别
	level := zapcore.InfoLevel
	if err := level.UnmarshalText([]byte(cfg.Level)); err != nil {
		return err
	}

	// 配置编码器
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

	// 配置输出
	var writeSyncer zapcore.WriteSyncer
	if cfg.Mode == "prod" {
		// 创建日志目录
		dir := filepath.Dir(cfg.FilePath)
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}

		// 生产模式: 输出到文件
		lumberjackLogger := &lumberjack.Logger{
			Filename:   cfg.FilePath,
			MaxSize:    cfg.MaxSize,
			MaxBackups: cfg.MaxBackups,
			MaxAge:     cfg.MaxAge,
			Compress:   cfg.Compress,
			LocalTime:  true, // 使用本地时间命名备份文件
		}
		writeSyncer = zapcore.AddSync(lumberjackLogger)
		// 生产模式默认 JSON 格式
		if !cfg.EncodeJson {
			cfg.EncodeJson = true
		}
	} else {
		// 开发模式输出到控制台
		writeSyncer = zapcore.AddSync(os.Stdout)
	}

	// 选择编码器
	var encoder zapcore.Encoder
	if cfg.EncodeJson {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// 构建 Logger 核心
	core := zapcore.NewCore(encoder, writeSyncer, level)

	// 配置选项
	options := []zap.Option{}
	if cfg.ShowLine {
		options = append(options,
			zap.AddCaller(),
			zap.AddCallerSkip(cfg.CallerSkip), // 跳过封装函数
			zap.AddStacktrace(zapcore.ErrorLevel),
		)
	}

	// 初始化全局 Logger
	GlobalLogger = zap.New(core, options...)
	SugarLogger = GlobalLogger.Sugar()

	// 测试日志
	GlobalLogger.Info("gofurry logger init success",
		String("mode", cfg.Mode),
		String("level", cfg.Level),
	)

	// 程序退出时自动刷写日志
	runtime.SetFinalizer(&GlobalLogger, func(l **zap.Logger) {
		_ = (*l).Sync()
	})

	return nil
}

// customTimeEncoder 自定义时间编码器
func customTimeEncoder(timeFormat string) zapcore.TimeEncoder {
	return func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString(t.Format(timeFormat))
	}
}

// Sync 刷写日志缓冲区
func Sync() error {
	if GlobalLogger != nil {
		return GlobalLogger.Sync()
	}
	return nil
}

// ============================ 扩展结构化字段 ============================
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

// 新增常用字段类型
func Bool(key string, value bool) zap.Field {
	return zap.Bool(key, value)
}

func Float64(key string, value float64) zap.Field {
	return zap.Float64(key, value)
}

func Any(key string, value interface{}) zap.Field {
	return zap.Any(key, value)
}

// ============================ 日志调用方法 ============================
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

// ============================ 格式化日志 ============================
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

// ============================ 带字段的结构化日志 ============================
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
