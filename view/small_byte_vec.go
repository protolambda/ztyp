package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

// To represent views of < 32 bytes efficiently as just a slice of those bytes.
type SmallByteVecMeta uint8

func (td SmallByteVecMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (td SmallByteVecMeta) ViewFromBacking(node Node, _ ViewHook) (View, error) {
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

type SmallByteVecView []byte

func (v SmallByteVecView) Backing() Node {
	out := &Root{}
	copy(out[:], v)
	return out
}

const Bytes4 SmallByteVecMeta = 4
