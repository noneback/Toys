# Bloom Filter
A Go implementation of [BloomFilter](https://en.wikipedia.org/wiki/Bloom_filter)

## Usage

```go
	bl := NewBloomFilter(uint64(4000), 7)
	cases1 := []struct {
		key []byte
		ans bool
	}{
		{[]byte("Test1"), true},
		{[]byte("Test2"), true},
		{[]byte("Test3"), true},
	}
	cases2 := []struct {
		key []byte
		ans bool
	}{
		{[]byte("Test1"), true},
		{[]byte("Test5"), false},
		{[]byte("Test3"), true},
	}

	for _, c := range cases1 {
		bl.AddKey(c.key)
	}

	for _, c := range cases2 {
		if res := bl.MayHasKey(c.key); res != c.ans {
			fmt.Printf("error, expected %v, got %v", c.ans, res)
		}
	}



```
