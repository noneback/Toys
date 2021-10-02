package skiplist

import (
	"errors"
	"fmt"
	"lsm/codec"
	"math/rand"
	"sync"
)

type Sequencer interface {
	Create()
	Read()
	Update()
	Delete()
}

const (
	defaultMaxLevel = 48
	defaultLevel    = 1
)

var (
	skLevelError = errors.New("level error")
)

type Comparable func(a, b interface{}) bool

type SkipListNode struct {
	data *codec.Entry
	// score    int
	nextPtrs []*SkipListNode
}

type SkipList struct {
	header, tail *SkipListNode
	level        int
	size         int
	rwmtx        *sync.RWMutex
}

type SkipListOption struct {
	MaxSize  int
	MaxLevel int
}

func NewSkipListNode(levelNum int, data *codec.Entry) *SkipListNode {
	sln := &SkipListNode{
		nextPtrs: make([]*SkipListNode, levelNum),
		data:     data,
	}
	return sln
}

func NewSkipList() *SkipList {
	sl := &SkipList{
		level:  defaultLevel,
		header: NewSkipListNode(defaultMaxLevel, nil),
		size:   0,
		rwmtx:  &sync.RWMutex{},
	}
	return sl
}

func (sl *SkipList) Insert(data *codec.Entry) *SkipListNode {
	sl.rwmtx.Lock()
	defer sl.rwmtx.Unlock()

	h := sl.header
	updateNode := make([]*SkipListNode, defaultMaxLevel) // shores the node before newNode
	// search form the top level
	for i := sl.level - 1; i >= 0; i-- {
		// loop: 1. current nextPtrs is not empty && data is small than inserted one, curData < insertedData
		for h.nextPtrs[i] != nil && h.nextPtrs[i].data.Less(data) {
			h = h.nextPtrs[i]
		}
		updateNode[i] = h
	}
	// choose level to insert
	lvl := sl.randomLevel()
	if lvl > sl.level {
		// Insert into higher level, we need to create header -> tail
		for i := sl.level; i < lvl; i++ {
			updateNode[i] = sl.header
		}
		sl.level = lvl
	}
	// insert after updatedNote
	n := NewSkipListNode(lvl, data)
	for i := 0; i < lvl; i++ {
		n.nextPtrs[i] = updateNode[i].nextPtrs[i]
		updateNode[i].nextPtrs[i] = n
	}
	sl.size++

	return n
}

// Read get the entry if exists
func (sl *SkipList) Read(key []byte) (*codec.Entry, bool) {
	return nil, false
}

// random level determines which level should insert a ptr
func (sl *SkipList) randomLevel() int {
	ans := 1
	for rand.Intn(2) == 0 && ans <= defaultMaxLevel {
		ans++
	}
	return ans
}

func (sl *SkipList) Delete(d *codec.Entry) bool {
	return true
}

func (sl *SkipList) Update(d *codec.Entry) bool {
	return true
}

// Show show skiplist in console
func (sl *SkipList) Show() {
	for i := 0; i < sl.level; i++ {
		h := sl.header
		fmt.Println("[level]", i)
		for h.nextPtrs[i] != nil {
			fmt.Printf("{%v} -> ", h.nextPtrs[i].data.Key)
			h = h.nextPtrs[i]
		}
		fmt.Printf("[end]\n")
	}
}

func (sl *SkipList) Size() int {
	return sl.size
}
