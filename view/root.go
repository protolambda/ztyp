package view

import (
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	. "github.com/protolambda/ztyp/tree"
)

type RootMeta uint8

var _ TypeDef[*RootView] = RootMeta(0)

func (m RootMeta) Mask() TypeDef[View] {
	return Mask[*RootView, RootMeta]{T: m}
}

func (m RootMeta) Default(hook BackingHook) *RootView {
	return &RootView{}
}

func (m RootMeta) IsFixedByteLength() bool {
	return true
}

func (m RootMeta) TypeByteLength() uint64 {
	return 32
}

func (m RootMeta) MinByteLength() uint64 {
	return 32
}

func (m RootMeta) MaxByteLength() uint64 {
	return 32
}

func (m RootMeta) Deserialize(dr *codec.DecodingReader) (*RootView, error) {
	v := RootView{}
	_, err := dr.Read(v[:])
	return &v, err
}

func (m RootMeta) Name() string {
	return "Root"
}

func (m RootMeta) String() string {
	return "Root"
}

func (RootMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (RootMeta) ViewFromBacking(node Node, _ BackingHook) (*RootView, error) {
	root, ok := node.(*Root)
	if !ok {
		return nil, fmt.Errorf("node is not a root: %v", node)
	} else {
		return (*RootView)(root), nil
	}
}

const RootType RootMeta = 0

type RootView Root

func AsRoot(v View, err error) (Root, error) {
	if err != nil {
		return Root{}, err
	}
	r, ok := v.(*RootView)
	if !ok {
		return Root{}, fmt.Errorf("not a root view: %v", v)
	}
	return Root(*r), nil
}

func (r *RootView) Type() TypeDef[View] {
	return RootType.Mask()
}

// Backing, a root can be used as a view representing itself.
func (r *RootView) Backing() Node {
	return (*Root)(r)
}

func (r *RootView) SetBacking(b Node) error {
	if br, ok := b.(*Root); ok {
		*r = RootView(*br)
		return nil
	} else {
		return NavigationError
	}
}

func (r *RootView) Copy() (View, error) {
	return r, nil
}

func (r *RootView) ValueByteLength() (uint64, error) {
	return 32, nil
}

func (r *RootView) Serialize(w *codec.EncodingWriter) error {
	return w.Write(r[:])
}

func (r *RootView) HashTreeRoot(h HashFn) Root {
	return Root(*r)
}

func (r *RootView) MarshalText() ([]byte, error) {
	return conv.BytesMarshalText(r[:])
}

func (r *RootView) UnmarshalText(text []byte) error {
	return conv.FixedBytesUnmarshalText(r[:], text)
}

func (r *RootView) String() string {
	return conv.BytesString(r[:])
}
