package bloomfilter

// Filter detect contain a key or not
type Filter interface {
}

// BloomFilter
type BloomFilter struct {
	hasher func(key []byte) uint64 // hasher func
	k      int                     // num of hash func, num of rehash times
	bits   uint64                  // num of bits
	bitset *BitSet
}

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

func (bf *BloomFilter) SetHasher(hasher Hasher) {
	bf.hasher = hasher
}

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

func (bf *BloomFilter) Clear() {}

func murmurHash() {}
