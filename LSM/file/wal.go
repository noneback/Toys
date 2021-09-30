package file

import (
	"bufio"
	"fmt"
	"io"
	"lsm/codec"
	"os"
	"sync"
)

type WAL struct {
	rwmtx  *sync.RWMutex
	f      *os.File
	writer io.Writer
}

type Option struct {
	ID      uint64
	Fname   string
	Dir     string
	Path    string
	Flag    int
	MaxSize int
}

func NewWAL(opt *Option) (*WAL, error) {
	f, err := os.OpenFile(opt.Path, opt.Flag, os.ModeAppend)
	if err != nil {
		return nil, fmt.Errorf("Wal open failed: %v", err)
	}
	w := &WAL{
		f:      f,
		rwmtx:  &sync.RWMutex{},
		writer: bufio.NewWriter(f),
	}
	return w, err
}

func (w *WAL) Close() error {
	if err := w.f.Close(); err != nil {
		return err
	}

	return nil
}

func (w *WAL) Append(entry *codec.Entry) (int, error) {
	// 只进行append entry 到文件
	w.rwmtx.Lock()
	defer w.rwmtx.Unlock()
	// TODO: max size validation
	return w.writer.Write(entry.Encode())
}
