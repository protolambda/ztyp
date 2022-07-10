package view

import (
	"fmt"
	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

type BasicVectorTypeDef[EV BasicView, ET BasicTypeDef[EV]] struct {
	ElemType     ET
	VectorLength uint64
	ComplexTypeBase
}

var _ TypeDef[*BasicVectorView[BasicView, BasicTypeDef[BasicView]]] = (*BasicVectorTypeDef[BasicView, BasicTypeDef[BasicView]])(nil)

func BasicVectorType[EV BasicView, ET BasicTypeDef[EV]](elemType ET, length uint64) *BasicVectorTypeDef[EV, ET] {
	size := length * elemType.TypeByteLength()
	return &BasicVectorTypeDef[EV, ET]{
		ElemType:     elemType,
		VectorLength: length,
		ComplexTypeBase: ComplexTypeBase{
			MinSize:     size,
			MaxSize:     size,
			Size:        size,
			IsFixedSize: true,
		},
	}
}

func (td *BasicVectorTypeDef[EV, ET]) Mask() TypeDef[View] {
	return Mask[*BasicVectorView[EV, ET], *BasicVectorTypeDef[EV, ET]]{T: td}
}

func (td *BasicVectorTypeDef[EV, ET]) FromElements(v ...EV) (*BasicVectorView[EV, ET], error) {
	length := uint64(len(v))
	if length > td.VectorLength {
		return nil, fmt.Errorf("expected no more than %d elements, got %d", td.VectorLength, length)
	}
	bottomNodes, err := td.ElemType.PackViews(v)
	if err != nil {
		return nil, err
	}
	depth := CoverDepth(td.BottomNodeLength())
	rootNode, _ := SubtreeFillToContents(bottomNodes, depth)
	listView, _ := td.ViewFromBacking(rootNode, nil)
	return listView, nil
}

func (td *BasicVectorTypeDef[EV, ET]) ElementType() ET {
	return td.ElemType
}

func (td *BasicVectorTypeDef[EV, ET]) Length() uint64 {
	return td.VectorLength
}

func (td *BasicVectorTypeDef[EV, ET]) DefaultNode() Node {
	depth := CoverDepth(td.BottomNodeLength())
	return SubtreeFillToDepth(&ZeroHashes[0], depth)
}

func (td *BasicVectorTypeDef[EV, ET]) ViewFromBacking(node Node, hook BackingHook) (*BasicVectorView[EV, ET], error) {
	depth := CoverDepth(td.BottomNodeLength())
	return &BasicVectorView[EV, ET]{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				Hook:        hook,
				BackingNode: node,
			},
			depth: depth,
		},
		BasicVectorTypeDef: td,
	}, nil
}

func (td *BasicVectorTypeDef[EV, ET]) ElementsPerBottomNode() uint64 {
	return 32 / td.ElemType.TypeByteLength()
}

func (td *BasicVectorTypeDef[EV, ET]) BottomNodeLength() uint64 {
	perNode := td.ElementsPerBottomNode()
	return (td.VectorLength + perNode - 1) / perNode
}

func (td *BasicVectorTypeDef[EV, ET]) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := td.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (td *BasicVectorTypeDef[EV, ET]) Default(hook BackingHook) *BasicVectorView[EV, ET] {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v
}

func (td *BasicVectorTypeDef[EV, ET]) New() *BasicVectorView[EV, ET] {
	return td.Default(nil)
}

func (td *BasicVectorTypeDef[EV, ET]) Deserialize(dr *codec.DecodingReader) (*BasicVectorView[EV, ET], error) {
	scope := dr.Scope()
	if td.Size != scope {
		return nil, fmt.Errorf("expected size %d does not match scope %d", td.Size, scope)
	}
	contents := make([]byte, scope, scope)
	if _, err := dr.Read(contents); err != nil {
		return nil, err
	}
	bottomNodes, err := BytesIntoNodes(contents)
	if err != nil {
		return nil, err
	}
	depth := CoverDepth(td.BottomNodeLength())
	rootNode, _ := SubtreeFillToContents(bottomNodes, depth)
	listView, _ := td.ViewFromBacking(rootNode, nil)
	return listView, nil
}

func (td *BasicVectorTypeDef[EV, ET]) String() string {
	return fmt.Sprintf("Vector[%s, %d]", td.ElemType.String(), td.VectorLength)
}

type BasicVectorView[EV BasicView, ET BasicTypeDef[EV]] struct {
	SubtreeView
	*BasicVectorTypeDef[EV, ET]
}

var _ View = (*BasicVectorView[BasicView, BasicTypeDef[BasicView]])(nil)

func AsBasicVector[EV BasicView, ET BasicTypeDef[EV]](v View, err error) (*BasicVectorView[EV, ET], error) {
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BasicVectorView[EV, ET])
	if !ok {
		return nil, fmt.Errorf("view is not a basic vector: %v", v)
	}
	return bv, nil
}

func (tv *BasicVectorView[EV, ET]) Type() TypeDef[View] {
	return tv.BasicVectorTypeDef.Mask()
}

func (tv *BasicVectorView[EV, ET]) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
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

func (tv *BasicVectorView[EV, ET]) Get(i uint64) (EV, error) {
	var out EV
	if i >= tv.VectorLength {
		return out, fmt.Errorf("basic vector has length %d, cannot get index %d", tv.VectorLength, i)
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return out, err
	}
	return tv.ElemType.BasicViewFromBacking(r, subIndex)
}

func (tv *BasicVectorView[EV, ET]) Set(i uint64, v EV) error {
	if i >= tv.VectorLength {
		return fmt.Errorf("cannot set item at element index %d, basic vector only has %d elements", i, tv.VectorLength)
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	return tv.SubtreeView.SetNode(bottomIndex, v.BackingFromBase(r, subIndex))
}

func (tv *BasicVectorView[EV, ET]) Copy() *BasicVectorView[EV, ET] {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy
}

func (tv *BasicVectorView[EV, ET]) Iter() ElemIter[EV, ET] {
	i := uint64(0)
	return ElemIterFn[EV, ET](func() (elem EV, elemType ET, ok bool, err error) {
		if i < tv.VectorLength {
			elem, err = tv.Get(i)
			ok = true
			elemType = tv.ElemType
			i += 1
			return
		} else {
			return
		}
	})
}

func (tv *BasicVectorView[EV, ET]) ReadonlyIter() ElemIter[EV, ET] {
	return basicElemReadonlyIter[EV, ET](tv.BackingNode, tv.VectorLength, tv.depth, tv.ElemType)
}

func (tv *BasicVectorView[EV, ET]) ValueByteLength() (uint64, error) {
	return tv.Size, nil
}

func (tv *BasicVectorView[EV, ET]) Serialize(w *codec.EncodingWriter) error {
	contents := make([]byte, tv.Size, tv.Size)
	if err := SubtreeIntoBytes(tv.BackingNode, tv.depth, tv.BottomNodeLength(), contents); err != nil {
		return err
	}
	return w.Write(contents)
}
