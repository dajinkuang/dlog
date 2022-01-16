package dlog

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/dajinkuang/util/glsutil"
	"github.com/dajinkuang/util/ordermaputil"
	"io"
	"path"
	"runtime"
	"time"

	"github.com/dajinkuang/util/iputil"
	"github.com/labstack/gommon/color"
)

// SetTopic 设置日志Topic，在main中修改
func SetTopic(topic string, absolutePath string) {
	if _dLogJSON != nil {
		_dLogJSON.Close()
		_dLogJSONError.Close()
	}
	dir := "/tmp/go/log"
	if len(absolutePath) > 0 {
		dir = absolutePath
	}
	file, err := NewFileBackend(dir, topic+".log_json_std")
	if err != nil {
		panic(err)
	}
	_dLogJSON = NewDLogJSON(file, topic)
	SetLogger(_dLogJSON)
	fileErrorAbove, err := NewFileBackend(dir, topic+".log_json_error")
	if err != nil {
		panic(err)
	}
	_dLogJSONError = NewDLogJSON(fileErrorAbove, topic)
	SetLoggerError(_dLogJSONError)
}

// _dLogJSON 可以打印任何级别的日志
var _dLogJSON *dLogJSON

// _dLogJSONError 只打印 ERROR FATAL 日志
var _dLogJSONError *dLogJSON

// GetDLogJSON 获取到 dLogJSON
func GetDLogJSON() *dLogJSON {
	if _dLogJSON == nil {
		SetTopic(defaultTopic, "")
	}
	return _dLogJSON
}

// GetDLogJSONError 获取到 _dLogJSONError
func GetDLogJSONError() *dLogJSON {
	if _dLogJSONError == nil {
		SetTopic(defaultTopic, "")
	}
	return _dLogJSONError
}

// dLogJSON dLog json 格式日志实现
type dLogJSON struct {
	prefix string
	level  Lvl
	output io.Writer
	levels []string
	color  *color.Color
	dw     *dLogWriter
}

// NewDLogJSON 新建一个dLogJSON
func NewDLogJSON(w io.WriteCloser, topic string) *dLogJSON {
	if len(topic) <= 0 {
		topic = defaultTopic
	}
	l := &dLogJSON{
		level:  INFO,
		prefix: topic,
		color:  color.New(),
	}
	l.initLevels()
	l.dw = NewDLogWriter(w)
	l.SetOutput(l.dw)
	l.SetLevel(INFO)
	return l
}

// With 向context中设置kv
func (dl *dLogJSON) With(ctx context.Context, kv ...interface{}) context.Context {
	om := FromContext(ctx)
	if om == nil {
		om = ordermaputil.NewOrderMap()
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	for i := 0; i < len(kv); i += 2 {
		om.Set(fmt.Sprintf("%v", kv[i]), kv[i+1])
	}
	return setContext(ctx, om)
}

// logJSON 打印json格式的日志。kv 应该是成对的 数据, 类似: name,张三,age,10,...
func (dl *dLogJSON) logJSON(ctxExternal context.Context, v Lvl, kv ...interface{}) (err error) {
	if v < dl.level {
		return nil
	}
	om := ordermaputil.NewOrderMap()
	_, file, line, _ := runtime.Caller(3)
	file = dl.getFilePath(file)
	om.Set("dlog_prefix", dl.Prefix())
	om.Set("level", dl.levels[v])
	now := time.Now()
	om.Set("cur_time", now.Format(time.RFC3339Nano))
	om.Set("cur_unix_time", now.Unix())
	om.Set("file", file)
	om.Set("line", line)
	localMachineIPV4, _ := iputil.LocalMachineIPV4()
	om.Set("local_machine_ipv4", localMachineIPV4)
	if ctxExternal == nil {
		ctxGls, ctxIsDefault := glsutil.GlsContext()
		if !ctxIsDefault {
			om.Set(TraceID, ValueFromOM(ctxGls, TraceID))
			om.Set(SpanID, ValueFromOM(ctxGls, SpanID))
			om.Set(ParentID, ValueFromOM(ctxGls, ParentID))
			om.Set(UserRequestIP, ValueFromOM(ctxGls, UserRequestIP))
			om.AddValues(FromContext(ctxGls))
		} else {
			traceID, pSpanID, spanID := glsutil.GetOpenTracingFromGls()
			om.Set(TraceID, traceID)
			om.Set(SpanID, spanID)
			om.Set(ParentID, pSpanID)
		}
	} else {
		om.Set(TraceID, ValueFromOM(ctxExternal, TraceID))
		om.Set(SpanID, ValueFromOM(ctxExternal, SpanID))
		om.Set(ParentID, ValueFromOM(ctxExternal, ParentID))
		om.Set(UserRequestIP, ValueFromOM(ctxExternal, UserRequestIP))
		om.AddValues(FromContext(ctxExternal))
	}
	if len(kv)%2 != 0 {
		kv = append(kv, "unknown")
	}
	for i := 0; i < len(kv); i += 2 {
		om.Set(fmt.Sprintf("%v", kv[i]), kv[i+1])
	}
	str, _ := json.Marshal(om)
	str = append(str, []byte("\n")...)
	_, err = dl.Output().Write(str)
	return
}

// Debug 打印debug日志
func (dl *dLogJSON) Debug(kv ...interface{}) {
	dl.logJSON(nil, DEBUG, kv...)
}

// Info 打印info日志
func (dl *dLogJSON) Info(kv ...interface{}) {
	dl.logJSON(nil, INFO, kv...)
}

// Warn 打印warn日志
func (dl *dLogJSON) Warn(kv ...interface{}) {
	dl.logJSON(nil, WARN, kv...)
}

// Error 打印error日志
func (dl *dLogJSON) Error(kv ...interface{}) {
	dl.logJSON(nil, ERROR, kv...)
}

// Fatal 打印fatal日志
func (dl *dLogJSON) Fatal(kv ...interface{}) {
	dl.logJSON(nil, ERROR, kv...)
}

// DebugContext 打印debug日志 context
func (dl *dLogJSON) DebugContext(ctx context.Context, kv ...interface{}) {
	dl.logJSON(ctx, DEBUG, kv...)
}

// InfoContext 打印info日志 context
func (dl *dLogJSON) InfoContext(ctx context.Context, kv ...interface{}) {
	dl.logJSON(ctx, INFO, kv...)
}

// WarnContext 打印warn日志 context
func (dl *dLogJSON) WarnContext(ctx context.Context, kv ...interface{}) {
	dl.logJSON(ctx, WARN, kv...)
}

// ErrorContext 打印error日志 context
func (dl *dLogJSON) ErrorContext(ctx context.Context, kv ...interface{}) {
	dl.logJSON(ctx, ERROR, kv...)
}

// FatalContext 打印fatal日志 context
func (dl *dLogJSON) FatalContext(ctx context.Context, kv ...interface{}) {
	dl.logJSON(ctx, ERROR, kv...)
}

func (dl *dLogJSON) getFilePath(file string) string {
	dir, base := path.Dir(file), path.Base(file)
	return path.Join(path.Base(dir), base)
}

// Close 关闭日志打印
func (dl *dLogJSON) Close() error {
	if dl.dw != nil {
		dl.dw.Close()
		dl.dw = nil
	}
	return nil
}

// EnableDebug 开启debug日志
func (dl *dLogJSON) EnableDebug(b bool) {
	if b {
		dl.SetLevel(DEBUG)
	} else {
		dl.SetLevel(INFO)
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

func (dl *dLogJSON) initLevels() {
	dl.levels = []string{
		"-",
		"DEBUG",
		"INFO",
		"WARN",
		"ERROR",
		"FATAL",
	}
}

// Prefix 获取日志prefix
func (dl *dLogJSON) Prefix() string {
	return dl.prefix
}

// SetPrefix 设置日志prefix
func (dl *dLogJSON) SetPrefix(p string) {
	dl.prefix = p
}

// Level 获取打印日志等级
func (dl *dLogJSON) Level() Lvl {
	return dl.level
}

// SetLevel 设置打印日志级别
func (dl *dLogJSON) SetLevel(v Lvl) {
	dl.level = v
}

// Output 获取writer
func (dl *dLogJSON) Output() io.Writer {
	return dl.output
}

// SetOutput 设置writer
func (dl *dLogJSON) SetOutput(w io.Writer) {
	dl.output = w
}

// Color 获得颜色
func (dl *dLogJSON) Color() *color.Color {
	return dl.color
}
