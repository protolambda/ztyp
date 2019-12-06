package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type SubtreeView struct {
	BackingNode Node
	depth       uint8
}

func (stv *SubtreeView) Get(i uint64) (Node, error) {
	g, err := ToGindex64(i, stv.depth)
	if err != nil {
		return nil, err
	}
	return stv.BackingNode.Getter(g)
}

func (stv *SubtreeView) Set(i uint64, node Node) error {
	g, err := ToGindex64(i, stv.depth)
	if err != nil {
		return err
	}
	s, err := stv.BackingNode.Setter(g)
	if err != nil {
		return err
	}
	stv.BackingNode = s(node)
	return nil
}

func (stv *SubtreeView) Backing() Node {
	return stv.BackingNode
}

// Copy over the roots at the bottom of the subtree from left to right into dest (until dest is full)
func (stv *SubtreeView) IntoBytes(dest []byte) error {
	copyChunk := func(i uint64, dest []byte) error {
		v, err := stv.Get(i)
		if err != nil {
			return err
		}
		r, ok := v.(*Root)
		if !ok {
			return fmt.Errorf("basic vector bottom node is not a root, at bottom node index %d", i)
		}
		copy(dest, r[:])
		return nil
	}
	endChunk := uint64(len(dest)) >> 5
	// copy over full chunks
	for i := uint64(0); i < endChunk; i++ {
		if err := copyChunk(i, dest[i<<5:(i+1)<<5]); err != nil {
			return err
		}
	}
	// copy over partial last chunk
	if endChunk<<5 != uint64(len(dest)) {
		if err := copyChunk(endChunk, dest[endChunk<<5:]); err != nil {
			return err
		}
	}
	return nil
}
