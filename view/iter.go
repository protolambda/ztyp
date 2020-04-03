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
			if depth != 0 {
				stack[0] = node
			}
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

func nodeReadonlyIter(node Node, length uint64, depth uint8) NodeIter {
	stack := make([]Node, depth, depth)

	i := uint64(0)
	rootIndex := uint64(0)
	return NodeIterFn(func() (chunk Node, ok bool, err error) {
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
			if depth != 0 {
				stack[0] = node
			}
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

		// Return the actual element
		return node, true, nil
	})
}

func elemReadonlyIter(node Node, length uint64, depth uint8, elemType TypeDef) ElemIter {
	nodeIter := nodeReadonlyIter(node, length, depth)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		node, ok, err := nodeIter.Next()
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
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
	length := uint64(len(fields))
	i := uint64(0)
	nodeIter := nodeReadonlyIter(node, length, depth)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		node, ok, err := nodeIter.Next()
		if err != nil {
			return nil, false, err
		}
		if !ok {
			return nil, false, nil
		}
		if i >= length {
			return nil, false, fmt.Errorf("node iter went too far, i: %d, length: %d", i, length)
		}
		el, err := fields[i].Type.ViewFromBacking(node, nil)
		if err != nil {
			return nil, false, err
		}
		i += 1
		// Return the actual element
		return el, true, nil
	})
}
