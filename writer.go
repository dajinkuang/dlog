package dlog

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/dajinkuang/errors"
)

const (
	bufLine = 1000 // 缓存一千行
)

type dLogWriter struct {
	w            io.WriteCloser
	buffer       chan string
	closeStartCh chan struct{}
	closeEndCh   chan struct{}
}

// NewDLogWriter 新建一个dLogWriter
func NewDLogWriter(w io.WriteCloser) *dLogWriter {
	ret := new(dLogWriter)
	ret.w = w
	ret.buffer = make(chan string, bufLine)
	ret.closeStartCh = make(chan struct{})
	ret.closeEndCh = make(chan struct{})
	go ret.realWrite()
	return ret
}

// Write 写操作
func (w dLogWriter) Write(p []byte) (n int, err error) {
	count := 0
	for {
		select {
		case <-w.closeEndCh: // 等到end的时候才真正不让写，也就是close开始的时候还是可以写的
			os.Stdout.WriteString(time.Now().String() + ",dLogWriter is closed\n")
			return 0, errors.New("dLogWriter_closed")
		case w.buffer <- string(p):
			return len(p), nil
		case <-time.After(time.Millisecond * 20):
			// 如果满了，记录下来
			count++
			str := fmt.Sprintf(time.Now().String()+",logWrite channel is full len=%v count=%d\n", len(w.buffer), count)
			os.Stdout.WriteString(str)
		}
	}
	return
}

// Close 关闭
func (w dLogWriter) Close() error {
	os.Stdout.WriteString(time.Now().String() + ",dLogWriter_close(w.closeStartCh)\n")
	close(w.closeStartCh)
	<-w.closeEndCh
	os.Stdout.WriteString(time.Now().String() + ",dLogWriter_<-w.closeEndCh\n")
	err := w.w.Close()
	os.Stdout.WriteString(time.Now().String() + ",dLogWriter_w.w.Close()\n")
	return err
}

func (w dLogWriter) realWrite() {
	for {
		select {
		case p := <-w.buffer:
			w.write([]byte(p))
		case <-w.closeStartCh: // 开始关闭，清空已经有的数据
			w.Flush()           // 这个时候还可以接收新的数据了
			close(w.closeEndCh) // 这个时候不接收新的数据了
			return
		}
	}
	return
}

// Flush 把当前有的数据都写进去，如果超过1s没有数据才算做清空了，但是最多等5秒
func (w dLogWriter) Flush() (err error) {
	ch := time.After(time.Second * 2)
	for {
		select {
		case <-time.After(time.Second * 1):
			// 等了1s还没有数据，就认为已经清空了
			return
		case <-ch:
			// 最多等2s，强制退出
			return
		case p := <-w.buffer:
			w.write([]byte(p))
		}
	}
	return
}

func (w dLogWriter) write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return w.w.Write(p)
}
