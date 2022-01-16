// Package dlog 日志打印包
package dlog

import (
	"context"
	"fmt"
	"io"
	"path"
	"runtime"

	"github.com/dajinkuang/util/glsutil"
	"github.com/dajinkuang/util/iputil"
	"github.com/labstack/gommon/log"
)

var _dLog *dLog

// GetDLog 获取普通DLog
func GetDLog() *dLog {
	if _dLog == nil {
		SetTopic(defaultTopic, "")
	}
	return _dLog
}

// dLog 普通dLog定义
type dLog struct {
	*log.Logger
	dw *dLogWriter
}

const defaultTopic = "default_topic"

const defaultHeader = `${prefix} ${level} ${time_rfc3339}`

// NewDLog 新建普通DLog
func NewDLog(w io.WriteCloser, topic string) *dLog {
	if len(topic) <= 0 {
		topic = defaultTopic
	}
	ret := &dLog{
		Logger: log.New(topic),
	}
	ret.dw = NewDLogWriter(w)
	ret.SetOutput(ret.dw)
	ret.SetHeader(defaultHeader)
	ret.SetLevel(log.INFO)
	ret.EnableColor()
	return ret
}

// logStr 打印log kv 应该是成对的数据。类似: name,张三,age,10,...
func (p *dLog) logStr(ctxExternal context.Context, kv ...interface{}) string {
	_, file, line, _ := runtime.Caller(3)
	file = p.getFilePath(file)
	localMachineIPV4, _ := iputil.LocalMachineIPV4()
	var pre []interface{}
	if ctxExternal == nil {
		ctxGls, _ := glsutil.GlsContext()
		pre = []interface{}{"local_machine_ipv4", localMachineIPV4, TraceID, ValueFromOM(ctxGls, TraceID),
			SpanID, ValueFromOM(ctxGls, SpanID), ParentID, ValueFromOM(ctxGls, ParentID), UserRequestIP, ValueFromOM(ctxGls, UserRequestIP)}
	} else {
		pre = []interface{}{"local_machine_ipv4", localMachineIPV4, TraceID, ValueFromOM(ctxExternal, TraceID),
			SpanID, ValueFromOM(ctxExternal, SpanID), ParentID, ValueFromOM(ctxExternal, ParentID), UserRequestIP, ValueFromOM(ctxExternal, UserRequestIP)}
	}
	kv = append(pre, kv...)
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	strFmt := "%s %d "
	args := []interface{}{file, line}
	for i := 0; i < len(kv); i += 2 {
		strFmt += "[%v=%+v]"
		args = append(args, kv[i], kv[i+1])
	}
	str := fmt.Sprintf(strFmt, args...)
	return str
}

// Debug 打印debug日志
func (p *dLog) Debug(kv ...interface{}) {
	p.Debugf("%s", p.logStr(nil, kv...))
}

// Info 打印info日志
func (p *dLog) Info(kv ...interface{}) {
	p.Infof("%s", p.logStr(nil, kv...))
}

// Warn 打印warn日志
func (p *dLog) Warn(kv ...interface{}) {
	p.Warnf("%s", p.logStr(nil, kv...))
}

// Error 打印error日志
func (p *dLog) Error(kv ...interface{}) {
	p.Errorf("%s", p.logStr(nil, kv...))
}

func (p *dLog) getFilePath(file string) string {
	dir, base := path.Dir(file), path.Base(file)
	return path.Join(path.Base(dir), base)
}

// Close 关闭打印日志
func (p *dLog) Close() error {
	if p.dw != nil {
		p.dw.Close()
		p.dw = nil
	}
	return nil
}

// DebugLog 开启debug日志
func (p *dLog) DebugLog(b bool) {
	if _dLog != nil {
		GetDLog().SetLevel(log.DEBUG)
	}
}

// DebugContext 打印debug日志 context
func (p *dLog) DebugContext(ctx context.Context, kv ...interface{}) {
	p.Debugf("%s", p.logStr(ctx, kv...))
}

// InfoContext 打印info日志 context
func (p *dLog) InfoContext(ctx context.Context, kv ...interface{}) {
	p.Infof("%s", p.logStr(ctx, kv...))
}

// WarnContext 打印warn日志 context
func (p *dLog) WarnContext(ctx context.Context, kv ...interface{}) {
	p.Warnf("%s", p.logStr(ctx, kv...))
}

// ErrorContext 打印error日志 context
func (p *dLog) ErrorContext(ctx context.Context, kv ...interface{}) {
	p.Errorf("%s", p.logStr(ctx, kv...))
}
