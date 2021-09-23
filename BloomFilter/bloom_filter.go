package bloomfilter

// Filter detect contain a key or not
type Filter interface {
	MayHasKey(key []byte) bool
	AddKey(key []byte)
}

// BloomFilter
type BloomFilter struct {
	hasher func(key []byte) uint64 // hasher func
	k      int                     // num of hash func, num of rehash times
	bits   uint64                  // num of bits
	bitset *BitSet
}

// NewBloomFilter make a bloom filter with num of bits and hash time
func NewBloomFilter(bits uint64, k int) *BloomFilter {
	bl := &BloomFilter{
		bits:   bits,
		k:      k,
		bitset: NewBitSet(bits),
		hasher: Hash,
	}
	return bl
}

// MayHasKey tell whether has this key
func (bf *BloomFilter) MayHasKey(key []byte) bool {
	hashCode := bf.hasher(key)
	// trick, rotate code, pretending rehash
	delta := hashCode>>17 | hashCode<<15 // rotate 17 bits

	for i := 0; i < bf.k; i++ {
		bitPos := hashCode % bf.bits
		if bf.bitset.Get(bitPos) == BitOne {
			return true
		}
		hashCode += delta
	}

	return false
}

// change hash func
func (bf *BloomFilter) SetHasher(hasher Hasher) {
	bf.hasher = hasher
}

// AddKey add key to bl
func (bf *BloomFilter) AddKey(key []byte) {
	hashCode := bf.hasher(key)
	// trick, rotate code, pretending rehash
	delta := hashCode>>17 | hashCode<<15 // rotate 17 bits
	for i := 0; i < bf.k; i++ {
		bitPos := hashCode % bf.bits
		bf.bitset.Set(bitPos, BitOne) // set them to one
		hashCode += delta
	}
}

func (bf *BloomFilter) Clear() {
	for _, b := range bf.bitset.bkt {
		b.rwlck.Lock()
		b.b = 0
		b.rwlck.Unlock()
	}
}
