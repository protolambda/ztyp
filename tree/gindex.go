package tree

import "fmt"

type Gindex interface {
	// Subtree returns the same gindex, but with the anchor moved one bit to the right, to represent the subtree position.
	Subtree() Gindex
	// Anchor of the gindex: same depth, but with position zeroed out.
	Anchor() Gindex
	// Left child gindex
	Left() Gindex
	// Right child gindex
	Right() Gindex
	// Parent gindex
	Parent() Gindex
	// Get the depth of the gindex
	Depth() uint64
}

// TODO: implement big int based gindex to automatically switch to whenever the uint64 is too small

type Gindex64 uint64

func (v Gindex64) Subtree() Gindex {
	anchor := Gindex64(1 << GetDepth(uint64(v)))
	return v ^ anchor | (anchor >> 1)
}

func (v Gindex64) Anchor() Gindex {
	return Gindex64(1 << GetDepth(uint64(v)))
}

func (v Gindex64) Left() Gindex {
	return v << 1
}

func (v Gindex64) Right() Gindex {
	return v << 1 | 1
}

func (v Gindex64) Parent() Gindex {
	return v >> 1
}

func (v Gindex64) Depth() uint64 {
	return uint64(GetDepth(uint64(v)))
}

func ToGindex64(index uint64, depth uint8) (Gindex64, error) {
	if depth >= 64 {
		return 0, fmt.Errorf("depth %d is too deep for Gindex64", depth)
	}
	anchor := uint64(1) << depth
	if index >= anchor {
		return 0, fmt.Errorf("index %d is larger than anchor %d derived from depth %d", index, anchor, depth)
	}
	return Gindex64(anchor | index), nil
}
