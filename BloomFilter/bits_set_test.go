package bloomfilter

import (
	"testing"
)

func TestBitSet(t *testing.T) {
	bs := NewBitSet(10000)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			continue
		}
		bs.Set(uint64(i), BitOne)
	}
	bs.PrintBits()
}
