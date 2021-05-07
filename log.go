package dlog

import "context"

// 对外提供统一接口，可自定义替换
// 默认使用dlog
type Logger interface {
	Debug(kv ...interface{})
	Info(kv ...interface{})
	Warn(kv ...interface{})
	Error(kv ...interface{})
	Fatal(kv ...interface{})
	With(ctx context.Context, kv ...interface{}) context.Context // 增量附加字段 以后的日志都会带上这个日志
	Close() error
	EnableDebug(b bool)
}

var _dLogger Logger

var __dLoggerErrorAbove Logger

func SetLogger(l Logger) {
	_dLogger = l
}

func GetLogger() Logger {
	if _dLogger == nil {
		SetLogger(GetJsonDLog())
	}
	return _dLogger
}

func SetLoggerErrorAbove(l Logger) {
	__dLoggerErrorAbove = l
}

func GetLoggerErrorAbove() Logger {
	if __dLoggerErrorAbove == nil {
		SetLogger(GetJsonDLogErrorAbove())
	}
	return __dLoggerErrorAbove
}

func Debug(kv ...interface{}) {
	GetLogger().Debug(kv...)
}

func Info(kv ...interface{}) {
	GetLogger().Info(kv...)
}

func Warn(kv ...interface{}) {
	GetLogger().Warn(kv...)
}

func Error(kv ...interface{}) {
	GetLogger().Error(kv...)
	GetLoggerErrorAbove().Error(kv...)
}

func Fatal(kv ...interface{}) {
	GetLogger().Error(kv...)
	GetLoggerErrorAbove().Error(kv...)
}

func With(ctx context.Context, kv ...interface{}) context.Context {
	return GetLogger().With(ctx, kv...)
}

// 这个方法以后不要用了，请使用Close()
func Flush() error {
	return Close()
}

func Close() error {
	GetLoggerErrorAbove().Close()
	return GetLogger().Close()
}

func EnableDebug(b bool) {
	GetLogger().EnableDebug(b)
}
