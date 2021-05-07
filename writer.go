package dlog

import (
	"fmt"
	"github.com/dajinkuang/errors"
	"io"
	"os"
	"time"
)

const (
	bufLine = 1000 //缓存一千行
)

type dlogWriter struct {
	w            io.WriteCloser
	buffer       chan string
	closeStartCh chan struct{}
	closeEndCh   chan struct{}
}

func NewDlogWriter(w io.WriteCloser) *dlogWriter {
	ret := new(dlogWriter)
	ret.w = w
	ret.buffer = make(chan string, bufLine)
	ret.closeStartCh = make(chan struct{})
	ret.closeEndCh = make(chan struct{})
	go ret.realWrite()
	return ret
}

func (w dlogWriter) Write(p []byte) (n int, err error) {
	count := 0
	for {
		select {
		case <-w.closeEndCh: //等到end的时候 才真正不让写，也就是close 开始的时候还是可以写的
			os.Stdout.WriteString(time.Now().String() + ",dlogWriter is closed\n")
			return 0, errors.New("dlogWriter_closed")
		case w.buffer <- string(p):
			return len(p), nil
		case <-time.After(time.Millisecond * 20):
			//如果满了,记录下来
			count++
			str := fmt.Sprintf(time.Now().String()+",logWrite channel is full len=%v count=%d\n", len(w.buffer), count)
			os.Stdout.WriteString(str)
		}
	}
	return
}

func (w dlogWriter) Close() error {
	os.Stdout.WriteString(time.Now().String() + ",dlogWriter_close(w.closeStartCh)\n")
	close(w.closeStartCh)
	<-w.closeEndCh
	os.Stdout.WriteString(time.Now().String() + ",dlogWriter_<-w.closeEndCh\n")
	err := w.w.Close()
	os.Stdout.WriteString(time.Now().String() + ",dlogWriter_w.w.Close()\n")
	return err
}

func (w dlogWriter) realWrite() {
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

// 把当前有的数据都写进去，如果超过1s没有数据才算做清空了,但是最多等5秒
func (w dlogWriter) Flush() (err error) {
	ch := time.After(time.Second * 2)
	for {
		select {
		case <-time.After(time.Second * 1):
			//等了1s 还没有数据，就认为已经清空了
			return
		case <-ch:
			//最多等2s,强制退出
			return
		case p := <-w.buffer:
			w.write([]byte(p))
		}
	}
	return
}

func (w dlogWriter) write(p []byte) (n int, err error) {
	os.Stdout.Write(p)
	return w.w.Write(p)
}
