package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type RootMeta uint8

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

func (r *RootView) ValueByteLength() uint64 {
	return 32
}

func (r *RootView) Serialize(w io.Writer) error {
	_, err := w.Write(r[:])
	return err
}

func (r *RootView) HashTreeRoot(h HashFn) Root {
	return Root(*r)
}
