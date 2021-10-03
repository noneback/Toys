package skiplist

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"lsm/codec"
	"math/rand"
	"sync"
	"testing"
	"time"
)

func encode(val uint32) *codec.Entry {
	bs := bytes.NewBuffer([]byte{})
	if err := binary.Write(bs, binary.LittleEndian, val); err != nil {
		panic(err)
	}
	e := &codec.Entry{
		Key:   bs.Bytes(),
		Value: bs.Bytes(),
	}
	return e
}

func TestInsert(t *testing.T) {
	sl := NewSkipList(nil)
	rand.Seed(time.Now().Unix())
	wg := sync.WaitGroup{}
	for i := 0; i < 200; i++ {
		val := uint32(rand.Int()%100000 + 1)
		go func(wg *sync.WaitGroup) {
			wg.Add(1)
			defer wg.Done()
			sl.Insert(encode(val))
		}(&wg)
	}
	wg.Wait()

	n := sl.header
	cnt := 0
	fmt.Printf("%+v\n", sl)
	a, b := encode(0), encode(100000)
	for n.nextPtrs[0] != nil {
		n = n.nextPtrs[0]
		cnt++
		b = n.data
		if !a.Less(b) {
			t.Errorf("error: not in order")
		}
		a = b
	}
}

func TestSkiplistShow(t *testing.T) {
	sl := NewSkipList(nil)
	rand.Seed(time.Now().Unix())
	for i := 0; i < 200; i++ {
		val := uint32(rand.Int() % 100000)
		sl.Insert(encode(val))
	}

	sl.Show()
}
