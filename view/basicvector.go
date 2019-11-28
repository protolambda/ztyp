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

func (td *BasicVectorTypeDef) ViewFromBacking(node Node) (View, error) {
	depth := GetDepth(td.BottomNodeLength())
	return &BasicVectorView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		BasicVectorTypeDef: td,
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

func (td *BasicVectorTypeDef) New() *BasicVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode())
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
	return tv.ElementType.SubViewFromBacking(r, subIndex), nil
}
func (tv *BasicVectorView) copyChunk(i uint64, offset uint8, dest []byte) error {
	v, err := tv.SubtreeView.Get(i)
	if err != nil {
		return err
	}
	r, ok := v.(*Root)
	if !ok {
		return fmt.Errorf("basic vector bottom node is not a root, at bottom node index %d", i)
	}
	copy(dest, r[offset:])
	return nil
}

func (tv *BasicVectorView) IntoBytes(skip uint64, dest []byte) error {
	startChunk, subStart := tv.TranslateIndex(skip)
	// copy over partial first chunk
	if subStart != 0 {
		if err := tv.copyChunk(startChunk, subStart, dest[startChunk<<5+uint64(subStart):]); err != nil {
			return err
		}
		startChunk += 1
	}
	endChunk, subEnd := tv.TranslateIndex(skip + uint64(len(dest)))
	// copy over full chunks
	for i := startChunk; i < endChunk; i++ {
		if err := tv.copyChunk(i, 0, dest[i<<5:(i+1)<<5]); err != nil {
			return err
		}
	}
	// copy over partial last chunk
	if subEnd != 0 {
		if err := tv.copyChunk(endChunk, 0, dest[endChunk<<5:endChunk<<5+uint64(subEnd)]); err != nil {
			return err
		}
	}
	return nil
}

func (tv *BasicVectorView) Set(i uint64, v SubView) error {
	if i >= tv.Length {
		return fmt.Errorf("cannot set item at element index %d, basic vector only has %d elements", i, tv.Length)
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	return tv.SubtreeView.Set(bottomIndex, v.BackingFromBase(r, subIndex))
}
