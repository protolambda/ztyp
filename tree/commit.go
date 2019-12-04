package tree

import (
	"errors"
	"fmt"
)

// An immutable (L, R) pair with a link to the holding node.
// If L or R changes, the link is used to bind a new (L, *R) or (*L, R) pair in the holding value.
type Commit struct {
	Value    Root
	Left     Node
	Right    Node
}

func NewCommit(a Node, b Node) *Commit {
	return &Commit{Left:  a, Right: b}
}

func (c *Commit) MerkleRoot(h HashFn) Root {
	if c.Value != (Root{}) {
		return c.Value
	}
	if c.Left == nil || c.Right == nil {
		panic("invalid state, cannot have left without right")
	}
	c.Value = h(c.Left.MerkleRoot(h), c.Right.MerkleRoot(h))
	return c.Value
}

func (c *Commit) RebindLeft(v Node) Node {
	return NewCommit(v, c.Right)
}

func (c *Commit) RebindRight(v Node) Node {
	return NewCommit(c.Left, v)
}

func SubtreeFillToDepth(bottom Node, depth uint8) Node {
	node := bottom
	for i := uint64(0); i < uint64(depth); i++ {
		node = NewCommit(node, node)
	}
	return node
}

func SubtreeFillToLength(bottom Node, depth uint8, length uint64) (Node, error) {
	anchor := uint64(1) << depth
	if length > anchor {
		return nil, errors.New("too many nodes")
	}
	if length == anchor {
		return SubtreeFillToDepth(bottom, depth), nil
	}
	if depth == 1 {
		if length > 1 {
			return NewCommit(bottom, bottom), nil
		} else {
			return NewCommit(bottom, &ZeroHashes[0]), nil
		}
	}
	pivot := anchor >> 1
	if length <= pivot {
		left, err := SubtreeFillToLength(bottom, depth-1, length)
		if err != nil {
			return nil, err
		}
		return NewCommit(left, &ZeroHashes[0]), nil
	} else {
		left := SubtreeFillToDepth(bottom, depth-1)
		right, err := SubtreeFillToLength(bottom, depth-1, length - pivot)
		if err != nil {
			return nil, err
		}
		return NewCommit(left, right), nil
	}
}

func SubtreeFillToContents(nodes []Node, depth uint8) (Node, error) {
	if len(nodes) == 0 {
		return nil, errors.New("no nodes to fill subtree with")
	}
	anchor := uint64(1) << depth
	if uint64(len(nodes)) > anchor {
		return nil, errors.New("too many nodes")
	}
	if depth == 0 {
		return nodes[0], nil
	}
	if depth == 1 {
		if len(nodes) > 1 {
			return NewCommit(nodes[0], nodes[1]), nil
		} else {
			return NewCommit(nodes[0], &ZeroHashes[0]), nil
		}
	}
	pivot := anchor >> 1
	if uint64(len(nodes)) <= pivot {
		left, err := SubtreeFillToContents(nodes, depth-1)
		if err != nil {
			return nil, err
		}
		return NewCommit(left, &ZeroHashes[0]), nil
	} else {
		left, err := SubtreeFillToContents(nodes[:pivot], depth-1)
		if err != nil {
			return nil, err
		}
		right, err := SubtreeFillToContents(nodes[pivot:], depth-1)
		if err != nil {
			return nil, err
		}
		return NewCommit(left, right), nil
	}
}

func (c *Commit) Getter(target uint64, depth uint8) (Node, error) {
	if depth == 0 {
		if target != 0 {
			return nil, fmt.Errorf("root depth 0 only has a single node at target 0, cannot Get %d", target)
		}
		return c, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.Left, nil
		}
		if target == 1 {
			return c.Right, nil
		}
		return nil, fmt.Errorf("depth 1 only has two nodes at target 0 and 1, cannot Get %d", target)
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		return c.Left.Getter(target, depth-1)
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		return c.Right.Getter(target&^pivot, depth-1)
	}
}

func (c *Commit) ExpandInto(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		if target != 0 {
			return nil, fmt.Errorf("root depth 0 only has a single node at target 0, cannot ExpandInto %d", target)
		}
		return Identity, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
		return nil, fmt.Errorf("depth 1 only has two nodes at target 0 and 1, cannot ExpandInto %d", target)
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if inner, err := c.Left.ExpandInto(target, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindLeft), nil
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if inner, err := c.Right.ExpandInto(target&^pivot, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindRight), nil
		}
	}
}

func (c *Commit) Setter(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		if target != 0 {
			return nil, fmt.Errorf("root depth 0 only has a single node at target 0, cannot Set %d", target)
		}
		return Identity, nil
	}
	if depth == 1 {
		if target == 0 {
			return c.RebindLeft, nil
		}
		if target == 1 {
			return c.RebindRight, nil
		}
		return nil, fmt.Errorf("depth 1 only has two nodes at target 0 and 1, cannot Set %d", target)
	}
	if pivot := uint64(1) << (depth - 1); target < pivot {
		if c.Left == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no left node", target, depth)
		}
		if inner, err := c.Left.Setter(target, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindLeft), nil
		}
	} else {
		if c.Right == nil {
			return nil, fmt.Errorf("cannot find node at target %v in depth %v: no right node", target, depth)
		}
		if inner, err := c.Right.Setter(target&^pivot, depth-1); err != nil {
			return nil, err
		} else {
			return Compose(inner, c.RebindRight), nil
		}
	}
}

// TODO: do we need a batching pattern, to not rebind branch by branch? Or is it sufficient to only create setters with reasonable scope?
