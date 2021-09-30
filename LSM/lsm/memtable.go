package lsm

import (
	"log"
	"lsm/file"
	"lsm/skiplist"
)

type memtable struct {
	wal *file.WAL
	sl  *skiplist.SkipList
}

type MemtableOption struct {
	WALOpt *file.Option
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

func (m *memtable) Set() {}
func (m *memtable) Get() {}
