package dlog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dajinkuang/util"
	utilGls "github.com/dajinkuang/util/gls"
	"io"
	"path"
	"runtime"
	"time"

	"github.com/labstack/gommon/color"
)

// 在main中修改
func SetTopic(topic string, absolutePath string) {
	// 考虑重入
	if _dJsonLog != nil {
		_dJsonLog.Close()
		__dJsonLogErrorAbove.Close()
	}
	dir := "/tmp/go/log"
	if len(absolutePath) > 0 {
		dir = absolutePath
	}
	file, err := NewFileBackend(dir, topic+".log_json_std")
	if err != nil {
		panic(err)
	}
	_dJsonLog = NewDJsonLog(file, topic)
	SetLogger(_dJsonLog)

	fileErrorAbove, err := NewFileBackend(dir, topic+".log_json_error_above")
	if err != nil {
		panic(err)
	}
	__dJsonLogErrorAbove = NewDJsonLog(fileErrorAbove, topic)
	SetLoggerErrorAbove(__dJsonLogErrorAbove)
}

var _dJsonLog *dJsonLog

// 只打印 ERROR FATAL 日志
var __dJsonLogErrorAbove *dJsonLog

func GetJsonDLog() *dJsonLog {
	if _dJsonLog == nil {
		SetTopic(defaultTopic, "")
	}
	return _dJsonLog
}

func GetJsonDLogErrorAbove() *dJsonLog {
	if __dJsonLogErrorAbove == nil {
		SetTopic(defaultTopic, "")
	}
	return __dJsonLogErrorAbove
}

type dJsonLog struct {
	prefix string
	level  Lvl
	output io.Writer
	levels []string
	color  *color.Color
	dw     *dlogWriter
}

func NewDJsonLog(w io.WriteCloser, topic string) *dJsonLog {
	if len(topic) <= 0 {
		topic = defaultTopic
	}
	l := &dJsonLog{
		level:  INFO,
		prefix: topic,
		color:  color.New(),
	}
	l.initLevels()
	l.dw = NewDlogWriter(w)
	l.SetOutput(l.dw)
	l.SetLevel(INFO)
	return l
}

func (p *dJsonLog) With(ctx context.Context, kv ...interface{}) context.Context {
	om := FromContext(ctx)
	if om == nil {
		om = NewOrderMap()
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	for i := 0; i < len(kv); i += 2 {
		om.Set(fmt.Sprintf("%v", kv[i]), kv[i+1])
	}
	return setContext(ctx, om)
}

// kv 应该是成对的 数据, 类似: name,张三,age,10,...
func (p *dJsonLog) logJson(v Lvl, kv ...interface{}) (err error) {
	if v < p.level {
		return nil
	}
	om := NewOrderMap()
	_, file, line, _ := runtime.Caller(3)
	file = p.getFilePath(file)
	om.Set("dlog_prefix", p.Prefix())
	om.Set("level", p.levels[v])
	now := time.Now()
	om.Set("cur_time", now.Format(time.RFC3339Nano))
	om.Set("cur_unix_time", now.Unix())
	om.Set("file", file)
	om.Set("line", line)
	localMachineIPV4, _ := util.LocalMachineIPV4()
	om.Set("local_machine_ipv4", localMachineIPV4)
	ctx, ctxIsDefault := utilGls.GlsContext()
	if !ctxIsDefault {
		om.Set(TraceID, ValueFromOM(ctx, TraceID))
		om.Set(SpanID, ValueFromOM(ctx, SpanID))
		om.Set(ParentID, ValueFromOM(ctx, ParentID))
		om.Set(UserRequestIP, ValueFromOM(ctx, UserRequestIP))
		om.AddVals(FromContext(ctx))
	} else {
		traceID, pSpanID, spanID := utilGls.GetOpenTracingFromGls()
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
	str = append(str, []byte("\n")...)
	_, err = p.Output().Write(str)
	return
}

func (p *dJsonLog) Debug(kv ...interface{}) {
	p.logJson(DEBUG, kv...)
}

func (p *dJsonLog) Info(kv ...interface{}) {
	p.logJson(INFO, kv...)
}

func (p *dJsonLog) Warn(kv ...interface{}) {
	p.logJson(WARN, kv...)
}

func (p *dJsonLog) Error(kv ...interface{}) {
	p.logJson(ERROR, kv...)
}

func (p *dJsonLog) Fatal(kv ...interface{}) {
	p.logJson(ERROR, kv...)
}

func (p *dJsonLog) getFilePath(file string) string {
	dir, base := path.Dir(file), path.Base(file)
	return path.Join(path.Base(dir), base)
}

func (p *dJsonLog) Close() error {
	if p.dw != nil {
		p.dw.Close()
		p.dw = nil
	}
	return nil
}

func (p *dJsonLog) EnableDebug(b bool) {
	if b {
		p.SetLevel(DEBUG)
	} else {
		p.SetLevel(INFO)
	}
}

type Lvl uint8

const (
	DEBUG Lvl = iota + 1
	INFO
	WARN
	ERROR
	FATAL
	OFF
)

func (l *dJsonLog) initLevels() {
	l.levels = []string{
		"-",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"FATAL",
	}
}

func (l *dJsonLog) Prefix() string {
	return l.prefix
}

func (l *dJsonLog) SetPrefix(p string) {
	l.prefix = p
}

func (l *dJsonLog) Level() Lvl {
	return l.level
}

func (l *dJsonLog) SetLevel(v Lvl) {
	l.level = v
}

func (l *dJsonLog) Output() io.Writer {
	return l.output
}

func (l *dJsonLog) SetOutput(w io.Writer) {
	l.output = w
}

func (l *dJsonLog) Color() *color.Color {
	return l.color
}
