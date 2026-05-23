package log

/*
 * @Desc: 日志服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"bytes"
	"fmt"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

const sFunctionName = "caller"
const sFunctionLine = "line"
const sFunctionEvent = "component"

func init() {
	logger.SetFormatter(&LoggerFormatter{})
	logger.SetLevel(logrus.InfoLevel)

	//writer, err := rotatelogs.New(
	//	logName+".%Y%m%d.log",
	//	rotatelogs.WithLinkName(logName),
	//	rotatelogs.WithRotationTime(24*time.Hour),
	//	rotatelogs.WithRotationCount(uint(env.GetServerConfig().Log.LogRotationCount)),
	//	//rotatelogs.WithMaxAge(time.Hour*24),
	//)
	//if err != nil {
	//	logger.Errorf("config local file business for logger error: %v", err)
	//} else {
	//	lfHook := lfshook.NewHook(lfshook.WriterMap{
	//		logrus.DebugLevel: writer,
	//		logrus.InfoLevel:  writer,
	//		logrus.WarnLevel:  writer,
	//		logrus.ErrorLevel: writer,
	//		logrus.FatalLevel: writer,
	//		logrus.PanicLevel: writer,
	//	}, // 分割日志样式
	//		//&logrus.TextFormatter{
	//		//	TimestampFormat: common.TIME_FORMAT_DATE,
	//		//	DisableColors: true,
	//		//	FullTimestamp: true,
	//		//	DisableSorting: true,
	//		//}
	//		&LoggerFormatter{},
	//	)
	//	logger.SetFormatter(&LoggerFormatter{})
	//	logger.AddHook(lfHook)
	//	logLevel := strings.ToLower(env.GetServerConfig().Log.LogLevel)
	//	switch logLevel {
	//	case "info":
	//		logger.SetLevel(logrus.InfoLevel)
	//	case "debug":
	//		logger.SetLevel(logrus.DebugLevel)
	//	case "warn":
	//		logger.SetLevel(logrus.WarnLevel)
	//	case "error":
	//		logger.SetLevel(logrus.ErrorLevel)
	//	}
	//}
}

type LoggerFormatter struct {
}

func (f *LoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	var b bytes.Buffer

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	level := formatLevel(entry.Level)
	component := formatFieldValue(entry.Data[sFunctionEvent])
	if component == "" {
		component = "collector"
	}

	b.WriteString(timestamp)
	b.WriteByte(' ')
	b.WriteString(fmt.Sprintf("%-5s", level))
	b.WriteByte(' ')
	b.WriteByte('[')
	b.WriteString(component)
	b.WriteString("] ")
	b.WriteString(entry.Message)

	caller := formatFieldValue(entry.Data[sFunctionName])
	line := formatFieldValue(entry.Data[sFunctionLine])
	if caller != "" {
		b.WriteString(" | caller=")
		b.WriteString(shortCaller(caller))
		if line != "" && line != "0" {
			b.WriteByte(':')
			b.WriteString(line)
		}
	}

	keys := make([]string, 0, len(entry.Data))
	for key := range entry.Data {
		if key == sFunctionName || key == sFunctionLine || key == sFunctionEvent {
			continue
		}
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		b.WriteByte(' ')
		b.WriteString(key)
		b.WriteByte('=')
		b.WriteString(quoteIfNeeded(formatFieldValue(entry.Data[key])))
	}

	b.WriteByte('\n')
	return b.Bytes(), nil
}

func WithFieldsMsg(fields map[string]interface{}, msg interface{}) {
	logWithFields(logrus.InfoLevel, fields, msg)
}

func InfoFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(logrus.InfoLevel, fields, msg)
}

func WarnFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(logrus.WarnLevel, fields, msg)
}

func ErrorFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(logrus.ErrorLevel, fields, msg)
}

func logWithFields(level logrus.Level, fields map[string]interface{}, msg interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(2)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}

	logFields := logrus.Fields{}
	for key, value := range buildCallerFields(functionName, line, "") {
		logFields[key] = value
	}
	for key, value := range fields {
		if key == sFunctionName || key == sFunctionLine {
			continue
		}
		logFields[key] = value
	}
	logger.WithFields(logFields).Log(level, msg)
}

func buildCallerFields(functionName string, line int, event string) logrus.Fields {
	if strings.TrimSpace(event) == "" {
		event = env.GetServerConfig().Server.AppName
	}
	return logrus.Fields{sFunctionName: functionName, sFunctionLine: line, sFunctionEvent: event}
}

func Error(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Error(msg...)
}

func Debug(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Debug(msg...)
}

func Warn(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Warn(msg...)
}

func Info(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Info(msg...)
}

func formatFieldValue(value interface{}) string {
	if value == nil {
		return ""
	}
	switch v := value.(type) {
	case time.Duration:
		return v.String()
	case fmt.Stringer:
		return v.String()
	default:
		return fmt.Sprint(v)
	}
}

func formatLevel(level logrus.Level) string {
	if level == logrus.WarnLevel {
		return "WARN"
	}
	return strings.ToUpper(level.String())
}

func shortCaller(caller string) string {
	if caller == "" {
		return ""
	}
	return path.Base(caller)
}

func quoteIfNeeded(value string) string {
	if value == "" {
		return `""`
	}
	if strings.IndexFunc(value, func(r rune) bool {
		return r == ' ' || r == '\t' || r == '\n' || r == '"' || r == '|'
	}) >= 0 {
		return strconv.Quote(value)
	}
	return value
}
