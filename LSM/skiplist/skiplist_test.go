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
	for i := 0; i < 10; i++ {
		val := &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		}
		sl.Insert(val)
	}

	sl.Show()
}

func TestFindPreNode(t *testing.T) {
	sl := NewSkipList(nil)
	insertedData := make([]*codec.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		insertedData = append(insertedData, &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		})
	}

	for _, v := range insertedData {
		sl.Insert(v)
	}
	var node *SkipListNode
	var ok bool
	if node, ok = sl.findPreNode(insertedData[1].Key); !ok || node.data != insertedData[0] {
		sl.Show()
		t.Fatalf("error trying Get %+v, get %+v", string(insertedData[1].Key), node)
		t.Fail()
	}
}

func TestSkipListGet(t *testing.T) {
	sl := NewSkipList(nil)
	insertedData := make([]*codec.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		insertedData = append(insertedData, &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		})
	}

	for _, v := range insertedData {
		sl.Insert(v)
	}

	for _, v := range insertedData {
		var e *codec.Entry
		var ok bool
		if e, ok = sl.Get(v.Key); !ok {
			sl.Show()
			t.Fatalf("No entry found key %v, expected %+v", string(v.Key), v.ToString())
			t.Fail()
		}
		t.Logf("Get entry %+v", e.ToString())
	}
}

func TestSkipListDelete(t *testing.T) {
	sl := NewSkipList(nil)
	insertedData := make([]*codec.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		insertedData = append(insertedData, &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		})
	}
	key := []byte("1")
	size := sl.Size()
	for _, v := range insertedData {
		sl.Insert(v)
	}

	if ok := sl.Delete(key); !ok {
		t.Fatalf("error deleted failed")
		t.Fail()
	}

	if _, ok := sl.Get(key); ok || size == sl.size+1 {
		t.Fatalf("error delete failed, not deleted")
		t.Fail()
	}
}

func TestSkipListUpdate(t *testing.T) {
	sl := NewSkipList(nil)
	insertedData := make([]*codec.Entry, 0, 10)
	for i := 0; i < 10; i++ {
		insertedData = append(insertedData, &codec.Entry{
			Key:   []byte{byte('0' + i)},
			Value: []byte{byte('0' + i)},
		})
	}
	key := []byte("1")
	val := []byte("test")
	for _, v := range insertedData {
		sl.Insert(v)
	}

	if ok := sl.Update(key, val); !ok {
		t.Fatalf("error update failed")
		t.Fail()
	}

	if node, ok := sl.Get(key); !ok || !bytes.Equal(val, node.Value) {
		t.Fatalf("error updated failed, not found key %v", key)
		t.Fail()
	}

}
