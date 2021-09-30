package lsm

import (
	"lsm/file"
	"lsm/skiplist"
)

type memtable struct {
	wal *file.WAL
	sl  *skiplist.SkipList
}

func Newmemtable() *memtable {
	return nil
}
