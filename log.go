package dlog

import (
	"context"
	"encoding/json"
	"fmt"
	"path"
	"runtime"
	"time"

	"github.com/dajinkuang/util/glsutil"
	"github.com/dajinkuang/util/iputil"
	"github.com/dajinkuang/util/ordermaputil"
	"github.com/dajinkuang/villa-go/log"
)

// Logger 对外提供统一接口，可自定义替换
// 默认使用dLogJSON
type Logger interface {
	Debug(kv ...interface{})
	Info(kv ...interface{})
	Warn(kv ...interface{})
	Error(kv ...interface{})
	Fatal(kv ...interface{})

	DebugContext(ctx context.Context, kv ...interface{})
	InfoContext(ctx context.Context, kv ...interface{})
	WarnContext(ctx context.Context, kv ...interface{})
	ErrorContext(ctx context.Context, kv ...interface{})
	FatalContext(ctx context.Context, kv ...interface{})

	With(ctx context.Context, kv ...interface{}) context.Context // 增量附加字段 以后的日志都会带上这个日志
	Close() error
	EnableDebug(b bool)
}

var _dLogger Logger

var __dLoggerError Logger

// SetLogger 设置Logger
func SetLogger(l Logger) {
	_dLogger = l
}

// GetLogger 获取Logger
func GetLogger() Logger {
	if _dLogger == nil {
		SetLogger(GetDLogJSON())
	}
	return _dLogger
}

// SetLoggerError 设置error以上级别的Logger
func SetLoggerError(l Logger) {
	__dLoggerError = l
}

// GetLoggerError 获取error以上级别的Logger
func GetLoggerError() Logger {
	if __dLoggerError == nil {
		SetLogger(GetDLogJSONError())
	}
	return __dLoggerError
}

// Debug 包调用，打印debug日志
func Debug(kv ...interface{}) {
	if logV2Open {
		log.Debug(logJSON(DEBUG, kv...))
		return
	}
	GetLogger().Debug(kv...)
}

// Info 包调用，打印info日志
func Info(kv ...interface{}) {
	if logV2Open {
		log.Info(logJSON(INFO, kv...))
		return
	}
	GetLogger().Info(kv...)
}

// Warn 包调用，打印warn日志
func Warn(kv ...interface{}) {
	if logV2Open {
		log.Warn(logJSON(WARN, kv...))
		return
	}
	GetLogger().Warn(kv...)
}

// Error 包调用，打印error日志
func Error(kv ...interface{}) {
	if logV2Open {
		log.Error(logJSON(ERROR, kv...))
		return
	}
	GetLogger().Error(kv...)
	GetLoggerError().Error(kv...)
}

// Fatal 包调用，打印fatal日志
func Fatal(kv ...interface{}) {
	if logV2Open {
		log.Fatal(logJSON(FATAL, kv...))
		return
	}
	GetLogger().Fatal(kv...)
	GetLoggerError().Fatal(kv...)
}

// DebugContext 包调用，打印debug日志，context
func DebugContext(ctx context.Context, kv ...interface{}) {
	if logV2Open {
		log.DebugContext(ctx, logJSON(DEBUG, kv...))
		return
	}
	GetLogger().DebugContext(ctx, kv...)
}

// InfoContext 包调用，打印info日志，context
func InfoContext(ctx context.Context, kv ...interface{}) {
	if logV2Open {
		log.InfoContext(ctx, logJSON(INFO, kv...))
		return
	}
	GetLogger().InfoContext(ctx, kv...)
}

// WarnContext 包调用，打印warn日志，context
func WarnContext(ctx context.Context, kv ...interface{}) {
	if logV2Open {
		log.WarnContext(ctx, logJSON(WARN, kv...))
		return
	}
	GetLogger().WarnContext(ctx, kv...)
}

// ErrorContext 包调用，打印error日志，context
func ErrorContext(ctx context.Context, kv ...interface{}) {
	if logV2Open {
		log.ErrorContext(ctx, logJSON(ERROR, kv...))
		return
	}
	GetLogger().ErrorContext(ctx, kv...)
	GetLoggerError().ErrorContext(ctx, kv...)
}

// FatalContext 包调用，打印fatal日志，context
func FatalContext(ctx context.Context, kv ...interface{}) {
	if logV2Open {
		log.FatalContext(ctx, logJSON(FATAL, kv...))
		return
	}
	GetLogger().FatalContext(ctx, kv...)
	GetLoggerError().FatalContext(ctx, kv...)
}

// With 向ctx设置kv
func With(ctx context.Context, kv ...interface{}) context.Context {
	if logV2Open {
		return ctx
	}
	return GetLogger().With(ctx, kv...)
}

// Flush 清空日志 这个方法以后不要用了，请使用Close()
func Flush() error {
	if logV2Open {
		log.Sync()
		return nil
	}
	return Close()
}

// Close 清空日志
func Close() error {
	if logV2Open {
		log.Sync()
		return nil
	}
	GetLoggerError().Close()
	return GetLogger().Close()
}

// EnableDebug debug开关
func EnableDebug(b bool) {
	if logV2Open {
		return
	}
	GetLogger().EnableDebug(b)
}

// logJSON 生成日志数据JSON字符串。kv 应该是成对的数据, 类似: name,张三,age,10,...
func logJSON(v Lvl, kv ...interface{}) string {
	om := ordermaputil.NewOrderMap()
	_, file, line, _ := runtime.Caller(3)
	file = getFilePath(file)
	om.Set("dlog_prefix", prefix)
	om.Set("level", logLevels[v])
	now := time.Now()
	om.Set("cur_time", now.Format(time.RFC3339Nano))
	om.Set("cur_unix_time", now.Unix())
	om.Set("file", file)
	om.Set("line", line)
	localMachineIPV4, _ := iputil.LocalMachineIPV4()
	om.Set("local_machine_ipv4", localMachineIPV4)
	ctx, ctxIsDefault := glsutil.GlsContext()
	if !ctxIsDefault {
		om.Set(TraceID, ValueFromOM(ctx, TraceID))
		om.Set(SpanID, ValueFromOM(ctx, SpanID))
		om.Set(ParentID, ValueFromOM(ctx, ParentID))
		om.Set(UserRequestIP, ValueFromOM(ctx, UserRequestIP))
		om.AddValues(FromContext(ctx))
	} else {
		traceID, pSpanID, spanID := glsutil.GetOpenTracingFromGls()
		om.Set(TraceID, traceID)
		om.Set(SpanID, spanID)
		om.Set(ParentID, pSpanID)
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	for i := 0; i < len(kv); i += 2 {
		om.Set(fmt.Sprintf("%v", kv[i]), kv[i+1])
	}
	str, _ := json.Marshal(om)
	//str = append(str, []byte("\n")...)
	return string(str)
}

var logLevels = []string{
	"-",
	"DEBUG",
	"INFO",
	"WARN",
	"ERROR",
	"FATAL",
}

func getFilePath(file string) string {
	dir, base := path.Dir(file), path.Base(file)
	return path.Join(path.Base(dir), base)
}

var (
	prefix    = "default"
	logV2Open bool // 如果为true，使用 "github.com/dajinkuang/villa-go/log" 打印日志
)

// SetTopicV2 设置日志Topic
func SetTopicV2(topic string, logV2Status bool) {
	prefix = topic
	logV2Open = logV2Status
}
