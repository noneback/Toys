package skiplist

import (
	"bytes"
	"errors"
	"fmt"
	"log"

	"lsm/codec"
	"math/rand"
	"sync"
)

type Sequencer interface {
	Get()
	Update()
	Delete()
}

const (
	defaultMaxLevel = 48
	defaultLevel    = 1
	defaultMaxSize  = 100
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
	maxSize      int
}

type SkipListOption struct {
	// MaxSize  int
	MaxLevel int
}

func NewSkipListNode(levelNum int, data *codec.Entry) *SkipListNode {
	sln := &SkipListNode{
		nextPtrs: make([]*SkipListNode, levelNum),
		data:     data,
	}
	return sln
}

func NewSkipList(opt *SkipListOption) *SkipList {
	maxLevel := defaultMaxLevel
	if opt != nil && opt.MaxLevel > 0 {
		maxLevel = opt.MaxLevel
	}

	sl := &SkipList{
		level:  defaultLevel,
		header: NewSkipListNode(maxLevel, nil),
		size:   0,
		rwmtx:  &sync.RWMutex{},
	}
	return sl
}

// findPreNode find the node before node.key
func (sl *SkipList) findPreNode(key []byte) (*SkipListNode, bool) {
	// from top to bottom
	for i := sl.level - 1; i >= 0; i-- {
		if node, ok := sl.findLevelPreNode(i, key); ok {
			return node, true
		}
	}
	return nil, false
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
func (sl *SkipList) Get(key []byte) (*codec.Entry, bool) {
	sl.rwmtx.RLock()
	defer sl.rwmtx.RUnlock()

	node, ok := sl.findPreNode(key)
	if !ok {
		return nil, false
	}
	return node.nextPtrs[0].data, true
}

// random level determines which level should insert a ptr
func (sl *SkipList) randomLevel() int {
	ans := 1
	for rand.Intn(2) == 0 && ans <= defaultMaxLevel {
		ans++
	}
	return ans
}

// findLevelPreNode find the node before key in level
func (sl *SkipList) findLevelPreNode(level int, key []byte) (*SkipListNode, bool) {
	h := sl.header
	for h.nextPtrs[level] != nil && bytes.Compare(h.nextPtrs[level].data.Key, key) != 1 {
		log.Printf("[findlevel] %+v\n", string(h.nextPtrs[level].data.Key))
		if bytes.Equal(h.nextPtrs[level].data.Key, key) {
			return h, true
		}
		h = h.nextPtrs[level]
	}
	return nil, false
}

// Delete delete node of key
func (sl *SkipList) Delete(key []byte) bool {
	sl.rwmtx.Lock()
	defer sl.rwmtx.Unlock()

	hasFound := false
	// TODO: remove all level ptr
	for i := 0; i < sl.level; i++ {
		if node, ok := sl.findLevelPreNode(i, key); ok {
			node.nextPtrs[i] = node.nextPtrs[i].nextPtrs[i]
			hasFound = true
		}
	}

	defer func() {
		if hasFound {
			sl.size--
		}
	}()
	return hasFound
}

func (sl *SkipList) Update(k, v []byte) bool {
	node, ok := sl.findPreNode(k)
	if !ok {
		return false
	}
	node.nextPtrs[0].data.Value = v
	return true
}

// Show show skiplist in console
func (sl *SkipList) Show() {
	for i := 0; i < sl.level; i++ {
		h := sl.header
		fmt.Println("[level]", i)
		for h.nextPtrs[i] != nil {
			fmt.Printf("{%v} -> ", h.nextPtrs[i].data.ToString())
			h = h.nextPtrs[i]
		}
		fmt.Printf("[end]\n")
	}
}

func (sl *SkipList) Size() int {
	return sl.size
}
