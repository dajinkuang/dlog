package dlog

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"
)

const (
	bufferSize    = 256 * 1024
	flushDuration = time.Second * 5
)

var _ io.WriteCloser = &FileBackend{}

// FileBackend 日志文件读写
type FileBackend struct {
	mu            sync.Mutex
	file          *os.File
	buffer        *bufio.Writer
	dir           string // directory for log files
	name          string
	filePath      string
	lastCheck     uint64
	flushDuration time.Duration
	closeCh       chan struct{}
}

// Write 写操作
func (p *FileBackend) Write(b []byte) (n int, err error) {
	p.mustFileExist()
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buffer.Write(b)
}

// Flush 刷到磁盘
func (p *FileBackend) Flush() error {
	p.mu.Lock()
	defer p.mu.Unlock()
	return p.buffer.Flush()
}

// Close 关闭文件读写
func (p *FileBackend) Close() error {
	close(p.closeCh)
	p.mu.Lock()
	defer p.mu.Unlock()
	p.buffer.Flush()
	p.file.Sync()
	return p.file.Close()
}

func (p *FileBackend) monitorFiles() {
	p.lastCheck = getLastCheck(time.Now())
	for range time.NewTicker(time.Second * 5).C {
		fileName := path.Join(p.dir, p.name)
		check := getLastCheck(time.Now())
		if p.lastCheck >= check {
			continue
		}
		p.mu.Lock()
		os.Rename(fileName, fileName+fmt.Sprintf(".%d", p.lastCheck))
		p.lastCheck = check
		newFile, err := os.OpenFile(fileName, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			panic(err)
		}
		p.buffer.Flush()
		p.file.Close()
		p.file = newFile
		p.buffer.Reset(p.file)
		p.mu.Unlock()
	}
}

func getLastCheck(now time.Time) uint64 {
	return uint64(now.Year())*1000000 + uint64(now.Month())*10000 + uint64(now.Day())*100 + uint64(now.Hour())
}

func (p *FileBackend) flushFile() {
	ticker := time.NewTicker(p.flushDuration)
	for {
		select {
		case <-ticker.C:
			p.Flush()
		case <-p.closeCh:
			return
		}
	}
}

func (p *FileBackend) mustFileExist() {
	timeStr := time.Now().Format(".2006010215")
	filePath := path.Join(p.dir, p.name+timeStr)
	if filePath == p.filePath {
		return
	}
	p.mu.Lock()
	newFile, err := os.OpenFile(filePath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		panic(err)
	}
	p.buffer.Flush()
	p.file.Close()
	p.file = newFile
	p.buffer.Reset(p.file)
	p.filePath = filePath
	p.mu.Unlock()
}

// NewFileBackend 新建一个FileBackend
func NewFileBackend(dir, name string) (*FileBackend, error) {
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, err
	}
	fb := new(FileBackend)
	fb.dir = dir
	fb.name = name
	fb.buffer = bufio.NewWriterSize(fb.file, bufferSize)
	fb.flushDuration = flushDuration
	fb.closeCh = make(chan struct{})
	fb.mustFileExist()
	go fb.flushFile()
	return fb, nil
}
