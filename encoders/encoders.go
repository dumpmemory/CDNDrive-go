package encoders

type Encoder interface {
	Decode(data []byte) ([]byte, error)
	Encode(data []byte) []byte
}
