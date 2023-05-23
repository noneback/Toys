package lsm

import (
	"log"
	"lsm/codec"
	"lsm/file"
	"lsm/skiplist"
)

type Memtable struct {
	wal *file.WAL
	sl  *skiplist.SkipList
}

type MemtableOption struct {
	WALOpt    *file.Option
	MaxKVSize int
}

func Newmemtable(opt *MemtableOption) *Memtable {
	wal, err := file.NewWAL(opt.WALOpt)
	if err != nil {
		log.Fatal("Create WAL failed", err)
	}
	return &Memtable{
		wal: wal,
		sl:  skiplist.NewSkipList(nil),
	}
}

func (m *Memtable) Set(e *codec.Entry) error {
	// write to WAL file
	if _, err := m.wal.Append(e); err != nil {
		return err
	}
	defer m.wal.Flush()
	// 2. write to skiplist
	m.sl.Insert(e)
	// TODO:whether need to turn into a immutable
	return nil
}

func (m *Memtable) Get(key []byte) (*codec.Entry, bool) {
	// Get from skiplist
	return m.sl.Get(key)
}

func (m *Memtable) Size() int {
	return m.sl.Size()
}

// TODO: recovery recover from the WAL file
func (m *Memtable) recovery() {}

func (m *Memtable) Close() {
	m.wal.Close()
}
