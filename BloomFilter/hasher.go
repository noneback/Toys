package bloomfilter

import (
	"crypto/md5"
)

type Hasher func(key []byte) uint64

func hex2Uint64(raw []byte) uint64 {
	ans := uint64(0)
	for _, v := range raw {
		ans += uint64(v)
	}
	return ans
}

func Hash(key []byte) uint64 {
	c := md5.New()
	c.Write(key)
	raw := c.Sum([]byte(""))
	return hex2Uint64(raw)
}
