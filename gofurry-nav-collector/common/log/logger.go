package log

/*
 * @Desc: 日志服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"runtime"
	"strings"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"github.com/sirupsen/logrus"
)

var logger = logrus.New()

const sFunctionName = "s-FunctionName"
const sFunctionLine = "s-FunctionLine"
const sFunctionEvent = "s-Event"

func init() {

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

func WithFieldsMsg(fields map[string]interface{}, msg interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	fields[sFunctionName] = functionName
	fields[sFunctionLine] = line
	logger.WithFields(fields).Info(msg)
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
	logger.WithFields(buildCallerFields(functionName, line, "")).Error(msg)
}

func Debug(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Debug(msg)
}

func Warn(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Warn(msg)
}

func Info(msg ...interface{}) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	logger.WithFields(buildCallerFields(functionName, line, "")).Info(msg)
}
