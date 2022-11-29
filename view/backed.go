package view

import . "github.com/protolambda/ztyp/tree"

type BackedView struct {
	Hook        BackingHook
	BackingNode Node
}

func (v *BackedView) HashTreeRoot(h HashFn) Root {
	return v.BackingNode.MerkleRoot(h)
}

func (v *BackedView) Backing() Node {
	return v.BackingNode
}

func (v *BackedView) SetBacking(b Node) error {
	v.BackingNode = b
	return v.Hook.PropagateChangeMaybe(b)
}

func (v *BackedView) SetHook(h BackingHook) {
	v.Hook = h
}
