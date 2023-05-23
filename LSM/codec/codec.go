package codec

import "encoding/json"

// type Encoder interface {
// 	Encode() []byte
// }

// type Decoder interface {
// 	Decode()
// }

type JSONCodec struct {
}

var (
	defaultCodec = JSONCodec{}
)

func (je *JSONCodec) Encode(data interface{}) ([]byte, error) {
	rawData, err := json.Marshal(data)
	if err != nil {
		return nil, err
	}
	return rawData, err
}

func (je *JSONCodec) Decode(rawData []byte, v interface{}) error {
	if err := json.Unmarshal(rawData, v); err != nil {
		return err
	}
	return nil
}

func GetJsonCodec() *JSONCodec {
	return &defaultCodec
}
