package file

import (
	"fmt"
	"lsm/utils"
	"os"
	"testing"
)

func TestWALAppend(t *testing.T) {

	wal, err := NewWAL(&Option{
		ID:   1,
		Path: "./test.wal",
		Flag: os.O_CREATE | os.O_APPEND,
	})

	if err != nil {
		t.Fatalf("error: create wal failed")
	}
	defer wal.Close()

	n, err := wal.Append(utils.MockEntry(1121))
	if err != nil {
		t.Fatalf("%v %v", n, err)
	}

	fmt.Println(n)

}
