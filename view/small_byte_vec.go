package view

import (
	"errors"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

// To represent views of < 32 bytes efficiently as just a slice of those bytes.
// Values above 32 are invalid.
// For 32, using Root ([32]byte as underlying type) is better.
type SmallByteVecMeta uint8

func (td SmallByteVecMeta) Default(_ BackingHook) View {
	return make(SmallByteVecView, td, td)
}

func (td SmallByteVecMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (td SmallByteVecMeta) ViewFromBacking(node Node, _ BackingHook) (View, error) {
	r, ok := node.(*Root)
	if !ok {
		return nil, fmt.Errorf("backing must be a root")
	}
	if td > 32 {
		return nil, fmt.Errorf("SmallByteVecMeta can only be used for values 0...32")
	}
	v := make(SmallByteVecView, td, td)
	copy(v, r[:])
	return v, nil
}

func (td SmallByteVecMeta) IsFixedByteLength() bool {
	return true
}

func (td SmallByteVecMeta) TypeByteLength() uint64 {
	return uint64(td)
}

func (td SmallByteVecMeta) MinByteLength() uint64 {
	return uint64(td)
}

func (td SmallByteVecMeta) MaxByteLength() uint64 {
	return uint64(td)
}

func (td SmallByteVecMeta) Deserialize(r io.Reader, scope uint64) (View, error) {
	if scope < uint64(td) {
		return nil, fmt.Errorf("scope of %d not enough for small byte vec of %d bytes", scope, td)
	}
	v := make(SmallByteVecView, td, td)
	_, err := r.Read(v)
	return v, err
}

func (td SmallByteVecMeta) String() string {
	return fmt.Sprintf("Vector[byte, %d]", td)
}

type SmallByteVecView []byte

func (v SmallByteVecView) SetBacking(b Node) error {
	return errors.New("cannot set backing of SmallByteVecView")
}

func (v SmallByteVecView) Backing() Node {
	out := &Root{}
	copy(out[:], v)
	return out
}

func (v SmallByteVecView) Copy() (View, error) {
	return v, nil
}

func (v SmallByteVecView) ValueByteLength() (uint64, error) {
	return uint64(len(v)), nil
}

func (v SmallByteVecView) Serialize(w io.Writer) error {
	_, err := w.Write(v)
	return err
}

func (v SmallByteVecView) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	copy(newRoot[:], v)
	return newRoot
}

func (v SmallByteVecView) Type() TypeDef {
	return SmallByteVecMeta(len(v))
}

const Bytes4Type SmallByteVecMeta = 4
const Bytes8Type SmallByteVecMeta = 8
const Bytes16Type SmallByteVecMeta = 16
