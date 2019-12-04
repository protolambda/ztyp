package tree

import "fmt"

type Root [32]byte

// Backing, a root can be used as a view representing itself.
func (r *Root) Backing() Node {
	return r
}

func (r *Root) Getter(target Gindex) (Node, error) {
	if !target.IsRoot() {
		return nil, fmt.Errorf("A Root does not have any child nodes to Get")
	}
	return r, nil
}

func (r *Root) Setter(target Gindex) (Link, error) {
	if !target.IsRoot() {
		return nil, fmt.Errorf("A Root does not have any child nodes to Set")
	}
	return Identity, nil
}

func (r *Root) ExpandInto(target Gindex) (Link, error) {
	if target.IsRoot() {
		return Identity, nil
	}
	depth := target.Depth()
	startC := &Commit{
		Left:  &ZeroHashes[depth-1],
		Right: &ZeroHashes[depth-1],
	}
	return startC.ExpandInto(target.Subtree())
}

func (r *Root) MerkleRoot(h HashFn) Root {
	if r == nil {
		return Root{}
	}
	return *r
}
