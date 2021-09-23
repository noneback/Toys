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
	// bs.PrintBits()
}

func TestBitGet(t *testing.T) {
	bs := NewBitSet(10000)
	for i := 0; i < 100; i++ {
		if i%2 == 0 {
			continue
		}
		bs.Set(uint64(i), BitOne)
	}
	for i := 0; i < 100; i++ {
		ans := bs.Get(uint64(i))
		if i%2 == 0 && ans != BitZero {
			t.Errorf("error, expected %v, got %v\n", BitZero, ans)
		}

		if i%2 == 1 && ans != BitOne {
			t.Errorf("error, expected %v, got %v\n", BitOne, ans)
		}

		bs.Set(uint64(i), BitOne)
	}
	// bs.PrintBits()
}
