package file

import (
	"bytes"
	"log"
	"lsm/utils"
	"os"
	"testing"
)

func TestWALAppend(t *testing.T) {
	opt := &Option{
		ID:   1,
		Path: "./test.wal",
		Flag: os.O_CREATE | os.O_RDWR | os.O_APPEND,
	}
	wal, err := NewWAL(opt)

	if err != nil {
		t.Fatalf("error: create wal failed")
	}
	defer wal.Close()
	e := utils.MockEntry(1)
	n, err := wal.Append(e)
	if err != nil {
		t.Fatalf("%v %v", n, err)
	}
	wal.Flush()
	raw, err := wal.Read(1)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(raw, e.Encode()) {
		log.Println(e.Encode())
		log.Println(raw)
		t.Fatal("not equal")
	}
}
