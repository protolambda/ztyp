package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

func basicElemReadonlyIter(node Node, length uint64, depth uint8, elemType BasicTypeDef) ElemIter {
	stack := make([]Node, depth, depth)

	i := uint64(0)
	// max 32 elements per bottom nodes, uint8 is safe.
	perNode := 32 / uint8(elemType.TypeByteLength())
	j := uint8(i) % perNode
	if j == 0 {
		j = perNode
	}
	var currentRoot *Root
	rootIndex := uint64(0)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		// done yet?
		if i >= length {
			return nil, false, nil
		}
		// in the middle of a node currently? finish that first
		if j < perNode {
			el, err := elemType.BasicViewFromBacking(currentRoot, j)
			j += 1
			if err != nil {
				return nil, false, err
			}
			return el, true, nil
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
				return nil, false, err
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
				return nil, false, err
			}
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node
		}

		// Get leaf node as a root
		r, leafIsRoot := node.(*Root)
		if !leafIsRoot {
			return nil, false, fmt.Errorf("expected leaf node %d to be a Root type", i)
		}
		// remember the root, we may need it for multiple subviews
		currentRoot = r

		// get the first subview
		el, err := elemType.BasicViewFromBacking(currentRoot, 0)
		if err != nil {
			return nil, false, err
		}
		// indicate that we have done one subview, and may need more to be read. Next one would be index 1, if any.
		j = 1
		// Return the actual element
		return el, true, nil
	})
}

func elemReadonlyIter(node Node, length uint64, depth uint8, elemType TypeDef) ElemIter {
	stack := make([]Node, depth, depth)

	i := uint64(0)
	rootIndex := uint64(0)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		// done yet?
		if i >= length {
			return nil, false, nil
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
				return nil, false, err
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
				return nil, false, err
			}
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node
		}

		el, err := elemType.ViewFromBacking(node, nil)
		if err != nil {
			return nil, false, err
		}
		// Return the actual element
		return el, true, nil
	})
}

func fieldReadonlyIter(node Node, depth uint8, fields []FieldDef) ElemIter {
	stack := make([]Node, depth, depth)

	i := uint64(0)
	length := uint64(len(fields))
	rootIndex := uint64(0)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		// done yet?
		if i >= length {
			return nil, false, nil
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
				return nil, false, err
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
				return nil, false, err
			}
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node
		}

		el, err := fields[rootIndex].Type.ViewFromBacking(node, nil)
		if err != nil {
			return nil, false, err
		}
		// Return the actual element
		return el, true, nil
	})
}
