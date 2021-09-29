package codec

type Encoder interface {
	Encode(data interface{}) []byte
}

type Decoder interface {
	Decode()
}
