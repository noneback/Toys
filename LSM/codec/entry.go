package codec

import "bytes"

type Entry struct {
	Key   []byte
	Value []byte
}

// Less means if self <= other
func (e *Entry) Less(a *Entry) bool {
	return bytes.Compare(e.Key, a.Key) < 1
}
