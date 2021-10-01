package file

import (
	"bufio"
	"fmt"
	"log"
	"lsm/codec"
	"os"
	"sync"
)

type WAL struct {
	rwmtx *sync.RWMutex
	f     *os.File
	rw    *bufio.ReadWriter
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
	f, err := os.OpenFile(opt.Path, opt.Flag, os.ModeAppend|os.ModePerm)
	if err != nil {
		return nil, fmt.Errorf("Wal open failed: %v", err)
	}
	w := &WAL{
		f:     f,
		rwmtx: &sync.RWMutex{},
		rw:    bufio.NewReadWriter(bufio.NewReader(f), bufio.NewWriter(f)),
	}
	return w, err
}

func (w *WAL) Close() error {
	if err := w.f.Close(); err != nil {
		return err
	}

	return nil
}

func (w *WAL) Flush() {
	if err := w.rw.Flush(); err != nil {
		log.Fatalf("flush error : %v", err)
	}
}

func (w *WAL) Append(entry *codec.Entry) (int, error) {
	// 只进行append entry 到文件
	w.f.Seek(0, 2)
	w.rwmtx.Lock()
	defer w.rwmtx.Unlock()
	// TODO: max size validation
	return w.rw.Write(append(entry.Encode()))
}

// Read read log[idx]
func (w *WAL) Read(idx int) ([]byte, error) {
	if _, err := w.f.Seek(0, 0); err != nil {
		return nil, err
	}
	raw := make([]byte, 0)
	var err error
	for i := 0; i < idx; i++ {
		raw, err = w.rw.ReadBytes('\n')
		// log.Println(raw)
		if err != nil {
			return nil, err
		}
	}
	return raw, nil
}
