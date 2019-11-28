package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type BitVectorTypeDef struct {
	BitLength uint64
}

func (td *BitVectorTypeDef) DefaultNode() Node {
	depth := GetDepth(td.BottomNodeLength())
	inner := &Commit{}
	inner.ExpandInplaceDepth(&ZeroHashes[0], depth)
	return inner
}

func (td *BitVectorTypeDef) ViewFromBacking(node Node) (View, error) {
	depth := GetDepth(td.BottomNodeLength())
	return &BitVectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		BitVectorTypeDef: td,
	}, nil
}

func (td *BitVectorTypeDef) BottomNodeLength() uint64 {
	return (td.BitLength + 0xff) >> 8
}

func (td *BitVectorTypeDef) New() *BitVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode())
	return v.(*BitVectorView)
}

func BitvectorType(length uint64) *BitVectorTypeDef {
	return &BitVectorTypeDef{
		BitLength: length,
	}
}

type BitVectorView struct {
	SubtreeView
	*BitVectorTypeDef
}

func (tv *BitVectorView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *BitVectorView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = i>>8, uint8(i)
	v, err := tv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("bitvector bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (tv *BitVectorView) Get(i uint64) (BoolView, error) {
	if i >= tv.BitLength {
		return false, fmt.Errorf("bitvector has bit length %d, cannot get bit index %d", tv.BitLength, i)
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return false, err
	}
	return BoolType.BoolViewFromBitfieldBacking(r, subIndex)
}

func (tv *BitVectorView) Set(i uint64, v BoolView) error {
	if i >= tv.BitLength {
		return fmt.Errorf("cannot set item at element index %d, bitvector only has %d bits", i, tv.BitLength)
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	return tv.SubtreeView.Set(bottomIndex, v.BackingFromBitfieldBase(r, subIndex))
}
