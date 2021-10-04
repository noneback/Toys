package codec

import (
	"bytes"
	"fmt"
	"log"
)

type Entry struct {
	Key   []byte `json:"key"`
	Value []byte `json:"value"`
}

// Less means if self <= other
func (e *Entry) Less(a *Entry) bool {
	return bytes.Compare(e.Key, a.Key) < 1
}

func (e *Entry) Encode() []byte {
	raw, err := defaultCodec.Encode(e)
	if err != nil {
		log.Fatal(err)
	}
	return append(raw, byte('\n'))
}

func (e *Entry) Decode(data []byte, v interface{}) error {
	return defaultCodec.Decode(data, v)
}

func (e *Entry) ToString() string {
	if e == nil {
		return ""
	}
	return fmt.Sprintf("[key : %v, value : %v]", string(e.Key), string(e.Value))
}
