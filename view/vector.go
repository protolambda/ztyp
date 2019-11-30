package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type VectorTypeDef struct {
	ElementType TypeDef
	Length      uint64
}

func (td *VectorTypeDef) DefaultNode() Node {
	depth := GetDepth(td.Length)
	inner := &Commit{}
	// The same node N times: the node is immutable, so re-use is safe.
	defaultNode := td.ElementType.DefaultNode()
	inner.ExpandInplaceDepth(defaultNode, depth)
	return inner
}

func (td *VectorTypeDef) ViewFromBacking(node Node, hook ViewHook) (View, error) {
	depth := GetDepth(td.Length)
	return &VectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		VectorTypeDef: td,
		ViewHook: hook,
	}, nil
}

func (td *VectorTypeDef) New(hook ViewHook) *VectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*VectorView)
}

func VectorType(elemType TypeDef, length uint64) *VectorTypeDef {
	return &VectorTypeDef{
		ElementType: elemType,
		Length:      length,
	}
}

type VectorView struct {
	SubtreeView
	*VectorTypeDef
	ViewHook
}

func (tv *VectorView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *VectorView) Get(i uint64) (View, error) {
	if i >= tv.VectorTypeDef.Length {
		return nil, fmt.Errorf("cannot get item at element index %d, vector only has %d elements", i, tv.VectorTypeDef.Length)
	}
	v, err := tv.SubtreeView.Get(i)
	if err != nil {
		return nil, err
	}
	return tv.VectorTypeDef.ElementType.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *VectorView) Set(i uint64, v View) error {
	if i >= tv.VectorTypeDef.Length {
		return fmt.Errorf("cannot set item at element index %d, vector only has %d elements", i, tv.VectorTypeDef.Length)
	}
	if err := tv.SubtreeView.Set(i, v.Backing()); err != nil {
		return err
	}
	return tv.PropagateChange(tv)
}

func (tv *VectorView) ItemHook(i uint64) ViewHook {
	return func(v View) error {
		return tv.Set(i, v)
	}
}
