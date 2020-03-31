package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type BasicVectorTypeDef struct {
	ElemType BasicTypeDef
	Len      uint64
	ComplexTypeBase
}

func BasicVectorType(name string, elemType BasicTypeDef, length uint64) *BasicVectorTypeDef {
	size := length * elemType.TypeByteLength()
	return &BasicVectorTypeDef{
		ElemType: elemType,
		Len:      length,
		ComplexTypeBase: ComplexTypeBase{
			TypeName: name,
			MinSize: size,
			MaxSize: size,
			Size: size,
			IsFixedSize: true,
		},
	}
}

func (td *BasicVectorTypeDef) ElementType() TypeDef {
	return td.ElemType
}

func (td *BasicVectorTypeDef) Length() uint64 {
	return td.Len
}

func (td *BasicVectorTypeDef) DefaultNode() Node {
	depth := CoverDepth(td.BottomNodeLength())
	return SubtreeFillToDepth(&ZeroHashes[0], depth)
}

func (td *BasicVectorTypeDef) ViewFromBacking(node Node, hook BackingHook) (View, error) {
	depth := CoverDepth(td.BottomNodeLength())
	return &BasicVectorView{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				ViewBase: ViewBase{
					TypeDef: td,
				},
				Hook: hook,
				BackingNode: node,
			},
			depth:       depth,
		},
		BasicVectorTypeDef: td,
	}, nil
}

func (td *BasicVectorTypeDef) ElementsPerBottomNode() uint64 {
	return 32 / td.ElemType.TypeByteLength()
}

func (td *BasicVectorTypeDef) BottomNodeLength() uint64 {
	perNode := td.ElementsPerBottomNode()
	return (td.Len + perNode - 1) / perNode
}

func (td *BasicVectorTypeDef) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := td.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (td *BasicVectorTypeDef) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *BasicVectorTypeDef) New(hook BackingHook) *BasicVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*BasicVectorView)
}

func (td *BasicVectorTypeDef) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *BasicVectorTypeDef) String() string {
	return fmt.Sprintf("Vector[%s, %d]", td.ElemType.Name(), td.Len)
}

type BasicVectorView struct {
	SubtreeView
	*BasicVectorTypeDef
}

func (tv *BasicVectorView) HashTreeRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *BasicVectorView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = tv.TranslateIndex(i)
	v, err := tv.SubtreeView.GetNode(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("basic vector bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (tv *BasicVectorView) Get(i uint64) (BasicView, error) {
	if i >= tv.Len {
		return nil, fmt.Errorf("basic vector has length %d, cannot get index %d", tv.Len, i)
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ElemType.BasicViewFromBacking(r, subIndex)
}

func (tv *BasicVectorView) Set(i uint64, v BasicView) error {
	if i >= tv.Len {
		return fmt.Errorf("cannot set item at element index %d, basic vector only has %d elements", i, tv.Len)
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	return tv.SubtreeView.SetNode(bottomIndex, v.BackingFromBase(r, subIndex))
}

func (tv *BasicVectorView) Copy() (View, error) {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy, nil
}

func (tv *BasicVectorView) ValueByteLength() uint64 {
	return tv.Size
}

func (tv *BasicVectorView) Serialize(w io.Writer) error {
	// TODO
	return nil
}

