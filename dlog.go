package dlog

import (
	"fmt"
	"github.com/dajinkuang/util"
	gls2 "github.com/dajinkuang/util/gls"
	"io"
	"path"
	"runtime"

	"github.com/labstack/gommon/log"
)

var _dLog *dLog

func GetDLog() *dLog {
	if _dLog == nil {
		SetTopic(defaultTopic, "")
	}
	return _dLog
}

type dLog struct {
	*log.Logger
	dw *dlogWriter
}

const defaultTopic = "default_topic"

// const defaultHeader = `${prefix} ${level} ${time_rfc3339} ${short_file} ${line}`
const defaultHeader = `${prefix} ${level} ${time_rfc3339}`

func NewDLog(w io.WriteCloser, topic string) *dLog {
	if len(topic) <= 0 {
		topic = defaultTopic
	}
	ret := &dLog{
		Logger: log.New(topic),
	}
	ret.dw = NewDlogWriter(w)
	ret.SetOutput(ret.dw)
	ret.SetHeader(defaultHeader)
	ret.SetLevel(log.INFO)
	// ret.SetLevel(log.DEBUG)
	ret.EnableColor()
	return ret
}

// kv 应该是成对的 数据, 类似: name,张三,age,10,...
func (p *dLog) logStr(kv ...interface{}) string {
	_, file, line, _ := runtime.Caller(3)
	file = p.getFilePath(file)
	localMachineIPV4, _ := util.LocalMachineIPV4()
	ctx, _ := gls2.GlsContext()
	pre := []interface{}{"local_machine_ipv4", localMachineIPV4, TraceID, ValueFromOM(ctx, TraceID),
		SpanID, ValueFromOM(ctx, SpanID), ParentID, ValueFromOM(ctx, ParentID), UserRequestIP, ValueFromOM(ctx, UserRequestIP)}
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

func (p *dLog) Debug(kv ...interface{}) {
	p.Debugf("", p.logStr(kv...))
}

func (p *dLog) Info(kv ...interface{}) {
	p.Infof("", p.logStr(kv...))
}

func (p *dLog) Warn(kv ...interface{}) {
	p.Warnf("", p.logStr(kv...))
}

func (p *dLog) Error(kv ...interface{}) {
	p.Errorf("", p.logStr(kv...))
}

func (p *dLog) getFilePath(file string) string {
	dir, base := path.Dir(file), path.Base(file)
	return path.Join(path.Base(dir), base)
}

func (p *dLog) Close() error {
	if p.dw != nil {
		p.dw.Close()
		p.dw = nil
	}
	return nil
}

func (p *dLog) DebugLog(b bool) {
	if _dLog != nil {
		GetDLog().SetLevel(log.DEBUG)
	}
}
