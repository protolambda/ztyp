package codec

import "io"

type Serializable interface {
	Serialize(w io.Writer) error
}

type Encodable interface {
	Encode() ([]byte, error)
}
