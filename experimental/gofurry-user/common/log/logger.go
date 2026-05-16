package log

/*
 * @Desc: 日志服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"bytes"
	"fmt"
	"runtime"
	"strings"

	"github.com/gofurry/gofurry-user/roof/env"
	"github.com/bytedance/sonic"
	"github.com/sirupsen/logrus"
)

var logName = env.GetServerConfig().Log.LogPath
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

func (l *LoggerFormatter) Format(entry *logrus.Entry) ([]byte, error) {
	b := &bytes.Buffer{}
	if entry.Buffer != nil {
		b = entry.Buffer
	} else {
		b = &bytes.Buffer{}
	}
	timestamp := entry.Time.Format("2005-01-02 15:04:05:05.999")
	funcName := entry.Data[sFunctionName]
	if funcName == nil {
		funcName = "F"
	}
	funcLine := entry.Data[sFunctionLine]
	if funcLine == nil {
		funcLine = 0
	}
	funcEvent := entry.Data[sFunctionEvent]
	if funcEvent == nil {
		funcEvent = env.GetServerConfig().Server.AppName
	}
	targetMap := make(map[string]any)
	for key, value := range entry.Data {
		if key == sFunctionName || key == sFunctionLine || key == sFunctionEvent {
			continue
		}
		targetMap[key] = value
	}
	newLog := ""
	dataInfo := ""
	if len(targetMap) > 0 {
		dataJson, _ := sonic.Marshal(targetMap)
		dataInfo = string(dataJson)
	}
	msg := entry.Message
	if len(msg) > env.GetServerConfig().Log.LogChokeLength {
		msg = msg[0:env.GetServerConfig().Log.LogChokeLength] + "...节流"
	}
	newLog = fmt.Sprintf("[%s] [%s] [%s -> %d] [%s -> %s] %s\n",
		funcEvent, timestamp, funcName, funcLine, entry.Level, msg, dataInfo)
	b.WriteString(newLog)
	return b.Bytes(), nil
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
