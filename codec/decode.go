package codec

import "io"

type Deserializable interface {
	Deserialize(r io.Reader, scope uint64) error
}

type Decodable interface {
	Decode(x []byte) error
}
