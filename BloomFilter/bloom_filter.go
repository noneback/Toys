package bloomfilter

// Filter detect contain a key or not
type Filter interface {
}

// BloomFilter
type BloomFilter struct {
	hasher Hasher // hasher func
}

func (bf *BloomFilter) MayHasKey(key *[]byte) bool {
	return true
}

func (bf *BloomFilter) SetHashFunc() {}
func (bf *BloomFilter) AddKey()      {}
func (bf *BloomFilter) Clear()       {}

func murmurHash() {}
