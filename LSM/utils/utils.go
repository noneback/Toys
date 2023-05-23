package utils

import (
	"bytes"
	"encoding/binary"
	"lsm/codec"
)

func MockEntry(val uint32) *codec.Entry {
	bs := bytes.NewBuffer([]byte{})
	if err := binary.Write(bs, binary.LittleEndian, val); err != nil {
		panic(err)
	}
	e := &codec.Entry{
		Key:   bs.Bytes(),
		Value: bs.Bytes(),
	}
	return e
}
