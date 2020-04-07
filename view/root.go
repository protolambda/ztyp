package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type RootMeta uint8

func (m RootMeta) Default(hook BackingHook) View {
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

func (m RootMeta) Deserialize(r io.Reader, scope uint64) (View, error) {
	if scope < 32 {
		return nil, fmt.Errorf("scope of %d not enough for root", scope)
	}
	v := RootView{}
	_, err := r.Read(v[:])
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

func (RootMeta) ViewFromBacking(node Node, _ BackingHook) (View, error) {
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

func (r *RootView) Type() TypeDef {
	return RootType
}

// Backing, a root can be used as a view representing itself.
func (r *RootView) Backing() Node {
	return (*Root)(r)
}

func (r *RootView) SetBacking(b Node) error {
	return NavigationError
}

func (r *RootView) Copy() (View, error) {
	return r, nil
}

func (r *RootView) ValueByteLength() (uint64, error) {
	return 32, nil
}

func (r *RootView) Serialize(w io.Writer) error {
	_, err := w.Write(r[:])
	return err
}

func (r *RootView) HashTreeRoot(h HashFn) Root {
	return Root(*r)
}
