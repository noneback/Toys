package lsm

import (
	"log"
	"lsm/codec"
	"lsm/file"
	"lsm/skiplist"
)

type memtable struct {
	wal *file.WAL
	sl  *skiplist.SkipList
}

type MemtableOption struct {
	WALOpt    *file.Option
	MaxKVSize int
}

func Newmemtable(opt *MemtableOption) *memtable {
	wal, err := file.NewWAL(opt.WALOpt)
	if err != nil {
		log.Fatal("Create WAL failed", err)
	}
	return &memtable{
		wal: wal,
		sl:  skiplist.NewSkipList(),
	}
}

func (m *memtable) Set(e *codec.Entry) error {
	// write to WAL file
	if _, err := m.wal.Append(e); err != nil {
		return err
	}
	// 2. write to skiplist
	m.sl.Insert(e)
	// TODO:whether need to turn into a immutable
	return nil
}

func (m *memtable) Get(key []byte) (*codec.Entry, bool) {
	// Get from skiplist
	return nil, false
}

func (m *memtable) Size() int {
	return m.sl.Size()
}

// TODO: recovery recover from the WAL file
func recovery()

func (m *memtable) Close() {
	m.wal.Close()
}
