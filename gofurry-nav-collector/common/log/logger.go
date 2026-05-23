package log

/*
 * @Desc: 日志服务
 * @author: 福狼
 * @version: v1.0.0
 */

import (
	"bytes"
	"fmt"
	"os"
	"path"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofurry/gofurry-nav-collector/roof/env"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const sFunctionName = "caller"
const sFunctionLine = "line"
const sFunctionEvent = "component"

var logger = newLogger()

func newLogger() *zap.Logger {
	core := &humanCore{
		enabler: zapcore.InfoLevel,
		ws:      zapcore.AddSync(os.Stdout),
		mu:      &sync.Mutex{},
	}
	return zap.New(core)
}

type humanCore struct {
	enabler zapcore.LevelEnabler
	ws      zapcore.WriteSyncer
	fields  []zap.Field
	mu      *sync.Mutex
}

func (c *humanCore) Enabled(level zapcore.Level) bool {
	return c.enabler.Enabled(level)
}

func (c *humanCore) With(fields []zap.Field) zapcore.Core {
	cloned := *c
	cloned.fields = append(append([]zap.Field{}, c.fields...), fields...)
	return &cloned
}

func (c *humanCore) Check(entry zapcore.Entry, checked *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if c.Enabled(entry.Level) {
		return checked.AddCore(entry, c)
	}
	return checked
}

func (c *humanCore) Write(entry zapcore.Entry, fields []zap.Field) error {
	allFields := make([]zap.Field, 0, len(c.fields)+len(fields))
	allFields = append(allFields, c.fields...)
	allFields = append(allFields, fields...)
	data := encodeFields(allFields)
	line := formatLine(entry, data)

	c.mu.Lock()
	defer c.mu.Unlock()
	_, err := c.ws.Write(line)
	return err
}

func (c *humanCore) Sync() error {
	return c.ws.Sync()
}

func encodeFields(fields []zap.Field) map[string]interface{} {
	encoder := zapcore.NewMapObjectEncoder()
	for _, field := range fields {
		field.AddTo(encoder)
	}
	return encoder.Fields
}

func formatLine(entry zapcore.Entry, fields map[string]interface{}) []byte {
	var b bytes.Buffer

	timestamp := entry.Time.Format("2006-01-02 15:04:05.000")
	component := formatFieldValue(fields[sFunctionEvent])
	if component == "" {
		component = "collector"
	}

	b.WriteString(timestamp)
	b.WriteByte(' ')
	b.WriteString(fmt.Sprintf("%-5s", formatLevel(entry.Level)))
	b.WriteByte(' ')
	b.WriteByte('[')
	b.WriteString(component)
	b.WriteString("] ")
	b.WriteString(entry.Message)

	caller := formatFieldValue(fields[sFunctionName])
	line := formatFieldValue(fields[sFunctionLine])
	if caller != "" {
		b.WriteString(" | caller=")
		b.WriteString(shortCaller(caller))
		if line != "" && line != "0" {
			b.WriteByte(':')
			b.WriteString(line)
		}
	}

	keys := make([]string, 0, len(fields))
	for key := range fields {
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
		b.WriteString(quoteIfNeeded(formatFieldValue(fields[key])))
	}

	b.WriteByte('\n')
	return b.Bytes()
}

func WithFieldsMsg(fields map[string]interface{}, msg interface{}) {
	logWithFields(zapcore.InfoLevel, fields, msg, 2)
}

func InfoFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(zapcore.InfoLevel, fields, msg, 2)
}

func WarnFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(zapcore.WarnLevel, fields, msg, 2)
}

func ErrorFields(fields map[string]interface{}, msg interface{}) {
	logWithFields(zapcore.ErrorLevel, fields, msg, 2)
}

func logWithFields(level zapcore.Level, fields map[string]interface{}, msg interface{}, callerDepth int) {
	line, functionName := callerInfo(callerDepth)
	logFields := callerFields(functionName, line, "")
	for key, value := range fields {
		if key == sFunctionName || key == sFunctionLine {
			continue
		}
		logFields[key] = value
	}
	write(level, fmt.Sprint(msg), logFields)
}

func Error(msg ...interface{}) {
	line, functionName := callerInfo(1)
	write(zapcore.ErrorLevel, fmt.Sprint(msg...), callerFields(functionName, line, ""))
}

func Debug(msg ...interface{}) {
	line, functionName := callerInfo(1)
	write(zapcore.DebugLevel, fmt.Sprint(msg...), callerFields(functionName, line, ""))
}

func Warn(msg ...interface{}) {
	line, functionName := callerInfo(1)
	write(zapcore.WarnLevel, fmt.Sprint(msg...), callerFields(functionName, line, ""))
}

func Info(msg ...interface{}) {
	line, functionName := callerInfo(1)
	write(zapcore.InfoLevel, fmt.Sprint(msg...), callerFields(functionName, line, ""))
}

func write(level zapcore.Level, msg string, fields map[string]interface{}) {
	checked := logger.Check(level, msg)
	if checked == nil {
		return
	}
	checked.Write(toZapFields(fields)...)
}

func callerInfo(depth int) (int, string) {
	line, functionName := 0, "???"
	pc, _, line, ok := runtime.Caller(depth + 1)
	if ok {
		functionName = runtime.FuncForPC(pc).Name()
	}
	return line, functionName
}

func callerFields(functionName string, line int, event string) map[string]interface{} {
	if strings.TrimSpace(event) == "" {
		event = env.GetServerConfig().Server.AppName
	}
	return map[string]interface{}{sFunctionName: functionName, sFunctionLine: line, sFunctionEvent: event}
}

func toZapFields(fields map[string]interface{}) []zap.Field {
	zapFields := make([]zap.Field, 0, len(fields))
	for key, value := range fields {
		switch v := value.(type) {
		case time.Duration:
			zapFields = append(zapFields, zap.String(key, v.String()))
		case fmt.Stringer:
			zapFields = append(zapFields, zap.String(key, v.String()))
		default:
			zapFields = append(zapFields, zap.Any(key, value))
		}
	}
	return zapFields
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

func formatLevel(level zapcore.Level) string {
	if level == zapcore.WarnLevel {
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
