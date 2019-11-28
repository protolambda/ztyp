package tree

import "fmt"

type Root [32]byte

// Backing, a root can be used as a view representing itself.
func (r *Root) Backing() Node {
	return r
}

func (r *Root) Getter(target uint64, depth uint8) (Node, error) {
	if depth != 0 {
		return nil, fmt.Errorf("A Root does not have any child nodes to Get")
	}
	return r, nil
}

func (r *Root) Setter(target uint64, depth uint8) (Link, error) {
	if depth != 0 {
		return nil, fmt.Errorf("A Root does not have any child nodes to Set")
	}
	return Identity, nil
}

func (r *Root) ExpandInto(target uint64, depth uint8) (Link, error) {
	if depth == 0 {
		return Identity, nil
	}
	startC := &Commit{
		Left:  &ZeroHashes[depth-1],
		Right: &ZeroHashes[depth-1],
	}
	return startC.ExpandInto(target, depth-1)
}

func (r *Root) MerkleRoot(h HashFn) Root {
	if r == nil {
		return Root{}
	}
	return *r
}
