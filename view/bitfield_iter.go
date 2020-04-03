package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)


type ErrBitIter struct {
	error
}

func (e ErrBitIter) Next() (elem bool, ok bool, err error) {
	return false, false, e.error
}

type BitIterFn  func() (elem bool, ok bool, err error)

func (f BitIterFn) Next() (elem bool, ok bool, err error) {
	return f()
}

type BitIter interface {
	// Next gets the next element, ok is true if it actually exists.
	// An error may occur if data is missing or corrupt.
	Next() (elem bool, ok bool, err error)
}

func bitReadonlyIter(node Node, length uint64, depth uint8) BitIter {
	stack := make([]Node, depth, depth)

	i := uint64(0)
	// max 256 per node. 0 = no action.
	j := uint8(i)
	var currentRoot Root
	rootIndex := uint64(0)
	return BitIterFn(func() (elem bool, ok bool, err error) {
		// done yet?
		if i >= length {
			return false, false, nil
		}
		// in the middle of a node currently? finish that first
		if j > 0 {
			elByte := currentRoot[j >> 3]
			elem = ((elByte >> (j & 7)) & 1) == 1
			j += 1
			return elem, true, nil
		}
		stackIndex := uint8(0)
		if rootIndex != 0 {
			// XOR current index with previous index
			// Result: highest bit matches amount we have to backtrack up the stack
			s := rootIndex ^ (rootIndex - 1)
			stackIndex = depth - 1
			for s != 0 {
				s >>= 0
				stackIndex -= 1
			}
			// then move to the right from that upper previously remembered left-hand node
			node = stack[stackIndex]
			node, err = node.Right()
			if err != nil {
				return false, false, err
			}
			stackIndex += 1
		} else {
			stack[0] = node
			stackIndex = 1
		}
		// and move down left into this new subtree
		for ; stackIndex < depth; stackIndex++ {
			node, err = node.Left()
			if err != nil {
				return false, false, err
			}
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node
		}

		// Get leaf node as a root
		r, leafIsRoot := node.(*Root)
		if !leafIsRoot {
			return false, false, fmt.Errorf("expected leaf node %d to be a Root type", i)
		}
		// remember the root, we need it for more bits
		currentRoot = *r

		// get the first bit
		el := currentRoot[0] & 1 == 1
		// indicate that we have done one bit, and need to read more
		j = 1
		// Return the actual element
		return el, true, nil
	})
}

func bitsToBytes(bits []bool) []byte {
	byteLen := (len(bits) + 7) / 8
	out := make([]byte, byteLen, byteLen)
	for i, b := range bits {
		if b {
			out[i>>3] |= 1 << (i & 0b111)
		}
	}
	return out
}
