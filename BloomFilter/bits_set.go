package bloomfilter

import (
	"fmt"
	"sync"
)

const (
	BitZero uint64 = iota
	BitOne
	ByteBits uint64 = 8
)

type bucket struct {
	b     uint8 // byte
	rwlck sync.RWMutex
}

type BitSet struct {
	bits uint64
	bkt  []bucket
}

func (b *BitSet) GetBits() uint64 {
	return b.bits
}

func (b *BitSet) PrintBits() {
	for _, v := range b.bkt {
		v.rwlck.RLock()
		fmt.Printf("[%v],", byte(v.b))
		v.rwlck.RUnlock()
	}
}

func NewBitSet(bits uint64) *BitSet {
	bs := &BitSet{
		bits: bits,
		bkt:  make([]bucket, bits/8+1),
	}

	return bs
}

func (b *BitSet) locate(pos uint64) (uint64, uint64) {
	return pos / ByteBits, pos % ByteBits
}

func (b *BitSet) Get(pos uint64) uint64 {
	idx, offset := b.locate(pos)
	b.bkt[idx].rwlck.RLock()
	defer b.bkt[idx].rwlck.RUnlock()

	mask := uint8(1) << offset
	return uint64(b.bkt[idx].b & mask)
}

func (b *BitSet) Set(pos uint64, val uint64) {
	idx, offset := b.locate(pos)
	b.bkt[idx].rwlck.Lock()
	defer b.bkt[idx].rwlck.Unlock()

	b.bkt[idx].b &= ^(uint8(1) << offset)
	if val == BitOne {
		b.bkt[idx].b |= uint8(1) << offset
	}
}
