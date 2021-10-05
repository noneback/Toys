package lsm

import (
	"bytes"
	"lsm/codec"
	"lsm/file"
	"os"
	"testing"
)

func TestMemtable(t *testing.T) {
	opt := &MemtableOption{
		WALOpt: &file.Option{
			ID:   1,
			Path: "./test.wal",
			Flag: os.O_CREATE | os.O_RDWR | os.O_APPEND,
		},
		MaxKVSize: 10000,
	}
	m := Newmemtable(opt)
	defer os.Remove(opt.WALOpt.Path)

	insertedData := make([]*codec.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		insertedData = append(insertedData, &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		})
	}
	for _, v := range insertedData {
		if err := m.Set(v); err != nil {
			t.Fatalf("error while set data to memtable")
			t.Fail()
		}
	}
	m.sl.Show()

	for _, v := range insertedData {
		if node, ok := m.Get(v.Key); !ok || !bytes.Equal(v.Value, node.Value) {
			t.Fatalf("error no such data in memtable, expected %v, got %v", v.ToString(), node.ToString())
			t.Fail()
		}
	}

}
