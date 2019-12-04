package tree

// A link is called to rebind a value, and retrieve the new root node.
type Link func(v Node) Node

func Identity(v Node) Node {
	return v
}

func Compose(inner Link, outer Link) Link {
	return func(v Node) Node {
		return outer(inner(v))
	}
}

// TODO: refactor these to use generalized indices as tree position.

// Node of a binary merkle tree
type Node interface {
	Getter(target Gindex) (Node, error)
	Setter(target Gindex) (Link, error)
	ExpandInto(target Gindex) (Link, error)
	MerkleRoot(h HashFn) Root
}
