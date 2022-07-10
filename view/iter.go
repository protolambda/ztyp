package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

func basicElemReadonlyIter[EV BasicView, ET BasicTypeDef[EV]](anchor Node, length uint64, depth uint8, elemType ET) ElemIter[EV, ET] {
	i := uint64(0)
	// max 32 elements per bottom nodes, uint8 is safe.
	perNode := 32 / uint8(elemType.TypeByteLength())
	j := perNode

	if limit := (uint64(1) << depth) * uint64(perNode); limit < length {
		return ErrElemIter[EV, ET]{fmt.Errorf("cannot handle iterate length %d bottom subviews (%d per node) in subtree of depth %d deep (limit %d subviews)", length, perNode, depth, limit)}
	}

	stack := make([]Node, depth, depth)
	var currentRoot *Root
	rootIndex := uint64(0)
	return ElemIterFn[EV, ET](func() (elem EV, typ ET, ok bool, err error) {
		// done yet?
		if i >= length {
			return
		}
		// in the middle of a node currently? finish that first
		if j < perNode {
			elem, err = elemType.BasicViewFromBacking(currentRoot, j)
			if err != nil {
				return
			}
			// progress how many nodes we have seen so far in this bottom chunk
			j += 1
			// progress the general element counter,
			i += 1
			return elem, elemType, true, nil
		}
		var node Node
		stackIndex := uint8(0)
		if rootIndex != 0 {
			// XOR current index with previous index
			// Result: highest bit matches amount we have to backtrack up the stack
			s := rootIndex ^ (rootIndex - 1)
			stackIndex = depth
			for s != 0 {
				s >>= 1
				stackIndex -= 1
			}
			// then move to the right from that upper previously remembered left-hand node
			node = stack[stackIndex]
			node, err = node.Right()
			if err != nil {
				return
			}
			stackIndex += 1
		} else {
			node = anchor
		}
		// and move down left into this new subtree
		for ; stackIndex < depth; stackIndex++ {
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node

			node, err = node.Left()
			if err != nil {
				return
			}
		}

		// Get leaf node as a root
		r, leafIsRoot := node.(*Root)
		if !leafIsRoot {
			err = fmt.Errorf("expected leaf node %d to be a Root type", i)
			return
		}
		// remember the root, we may need it for multiple subviews
		currentRoot = r

		// get the first subview
		elem, err = elemType.BasicViewFromBacking(currentRoot, 0)
		if err != nil {
			return
		}
		// indicate that we have done one subview, and may need more to be read. Next one would be index 1, if any.
		j = 1
		// And that the next root will be 1 after
		rootIndex += 1
		// And we progress the general element counter
		i += 1

		// Return the actual element
		return elem, elemType, true, nil
	})
}

func nodeReadonlyIter(anchor Node, length uint64, depth uint8) NodeIter {
	if limit := uint64(1) << depth; limit < length {
		return ErrNodeIter{fmt.Errorf("cannot handle iterate length %d nodes in subtree of depth %d deep (limit %d)", length, depth, limit)}
	}
	stack := make([]Node, depth, depth)

	i := uint64(0)
	return NodeIterFn(func() (chunk Node, ok bool, err error) {
		// done yet?
		if i >= length {
			return nil, false, nil
		}
		var node Node
		stackIndex := uint8(0)
		if i != 0 {
			// XOR current index with previous index
			// Result: highest bit matches amount we have to backtrack up the stack
			s := i ^ (i - 1)
			stackIndex = depth
			for s != 0 {
				s >>= 1
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
			node = anchor
		}
		// and move down left into this new subtree
		for ; stackIndex < depth; stackIndex++ {
			// remember left-hand nodes, we may revisit them
			stack[stackIndex] = node

			node, err = node.Left()
			if err != nil {
				return nil, false, err
			}
		}
		i += 1

		// Return the actual element
		return node, true, nil
	})
}

func elemReadonlyIter[EV View, ET TypeDef[EV]](node Node, length uint64, depth uint8, elemType ET) ElemIter[EV, ET] {
	nodeIter := nodeReadonlyIter(node, length, depth)
	// Re-use a typed view by changing its backing each iteration step.
	elemView := elemType.Default(nil)
	return ElemIterFn[EV, ET](func() (elem EV, typ ET, ok bool, err error) {
		var nodeEl Node
		nodeEl, ok, err = nodeIter.Next()
		if err != nil || !ok {
			return
		}
		if berr := elemView.SetBacking(nodeEl); berr != nil {
			// Create a new view if we can't re-use the existing view.
			elem, err = elemType.ViewFromBacking(nodeEl, nil)
			if err != nil {
				return
			}
			// Return the actual element
			return elem, elemType, true, nil
		} else {
			return elemView, elemType, true, nil
		}
	})
}

func fieldReadonlyIter(node Node, depth uint8, fields []FieldDef) ElemIter[View, TypeDef[View]] {
	length := uint64(len(fields))
	i := uint64(0)
	nodeIter := nodeReadonlyIter(node, length, depth)
	return ElemIterFn[View, TypeDef[View]](func() (elem View, elemType TypeDef[View], ok bool, err error) {
		node, ok, err := nodeIter.Next()
		if err != nil {
			return nil, nil, false, err
		}
		if !ok {
			return nil, nil, false, nil
		}
		if i >= length {
			return nil, nil, false, fmt.Errorf("node iter went too far, i: %d, length: %d", i, length)
		}
		elType := fields[i].Type
		el, err := elType.ViewFromBacking(node, nil)
		if err != nil {
			return nil, nil, false, err
		}
		i += 1
		// Return the actual element
		return el, elType, true, nil
	})
}
