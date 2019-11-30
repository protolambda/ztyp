package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type BasicVectorTypeDef struct {
	ElementType BasicTypeDef
	Length      uint64
}

func (td *BasicVectorTypeDef) DefaultNode() Node {
	depth := GetDepth(td.BottomNodeLength())
	inner := &Commit{}
	inner.ExpandInplaceDepth(&ZeroHashes[0], depth)
	return inner
}

func (td *BasicVectorTypeDef) ViewFromBacking(node Node, hook ViewHook) (View, error) {
	depth := GetDepth(td.BottomNodeLength())
	return &BasicVectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		BasicVectorTypeDef: td,
		ViewHook: hook,
	}, nil
}

func (td *BasicVectorTypeDef) ElementsPerBottomNode() uint64 {
	return 32 / td.ElementType.ByteLength()
}

func (td *BasicVectorTypeDef) BottomNodeLength() uint64 {
	perNode := td.ElementsPerBottomNode()
	return (td.Length + perNode - 1) / perNode
}

func (td *BasicVectorTypeDef) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := td.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (td *BasicVectorTypeDef) New(hook ViewHook) *BasicVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*BasicVectorView)
}

func BasicVectorType(elemType BasicTypeDef, length uint64) *BasicVectorTypeDef {
	return &BasicVectorTypeDef{
		ElementType: elemType,
		Length:      length,
	}
}

type BasicVectorView struct {
	SubtreeView
	*BasicVectorTypeDef
	ViewHook
}

func (tv *BasicVectorView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *BasicVectorView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = tv.TranslateIndex(i)
	v, err := tv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("basic vector bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (tv *BasicVectorView) Get(i uint64) (SubView, error) {
	if i >= tv.Length {
		return nil, fmt.Errorf("basic vector has length %d, cannot get index %d", tv.Length, i)
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ElementType.SubViewFromBacking(r, subIndex)
}

func (tv *BasicVectorView) Set(i uint64, v SubView) error {
	if i >= tv.Length {
		return fmt.Errorf("cannot set item at element index %d, basic vector only has %d elements", i, tv.Length)
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	if err := tv.SubtreeView.Set(bottomIndex, v.BackingFromBase(r, subIndex)); err != nil {
		return err
	}
	return tv.PropagateChange(tv)
}
