package bloomfilter

import (
	"crypto/md5"
)

type Hasher func(key string) int64

func Hash(key string) uint64 {
	buckets := 1232
	c := md5.New()
	c.Write([]byte(key))
	raw := c.Sum([]byte(""))
	return hex2Uint64(raw) % uint64(buckets)
}

func hex2Uint64(raw []byte) uint64 {
	ans := uint64(0)
	for _, v := range raw {
		ans += uint64(v)
	}
	return ans
}
