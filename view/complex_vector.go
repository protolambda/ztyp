package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type ComplexVectorTypeDef struct {
	ElemType     TypeDef
	VectorLength uint64
	ComplexTypeBase
}

func ComplexVectorType(name string, elemType TypeDef, length uint64) *ComplexVectorTypeDef {
	minSize := uint64(0)
	maxSize := uint64(0)
	size := uint64(0)
	isFixedSize := elemType.IsFixedByteLength()
	if isFixedSize {
		size = length * elemType.TypeByteLength()
		minSize = size
		maxSize = size
	} else {
		minSize = (length + OffsetByteLength) * elemType.MinByteLength()
		maxSize = (length + OffsetByteLength) * elemType.MaxByteLength()
	}
	return &ComplexVectorTypeDef{
		ElemType:     elemType,
		VectorLength: length,
		ComplexTypeBase: ComplexTypeBase{
			TypeName: name,
			MinSize: minSize,
			MaxSize: maxSize,
			Size: size,
			IsFixedSize: isFixedSize,
		},
	}
}

func (td *ComplexVectorTypeDef) ElementType() TypeDef {
	return td.ElemType
}

func (td *ComplexVectorTypeDef) Length() uint64 {
	return td.VectorLength
}

func (td *ComplexVectorTypeDef) DefaultNode() Node {
	depth := CoverDepth(td.VectorLength)
	// The same node N times: the node is immutable, so re-use is safe.
	defaultNode := td.ElemType.DefaultNode()
	// can ignore error, depth is derived from length.
	rootNode, _ := SubtreeFillToLength(defaultNode, depth, td.VectorLength)
	return rootNode
}

func (td *ComplexVectorTypeDef) ViewFromBacking(node Node, hook BackingHook) (View, error) {
	depth := CoverDepth(td.VectorLength)
	return &ComplexVectorView{
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
		ComplexVectorTypeDef: td,
	}, nil
}

func (td *ComplexVectorTypeDef) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *ComplexVectorTypeDef) New(hook BackingHook) *ComplexVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*ComplexVectorView)
}

func (td *ComplexVectorTypeDef) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *ComplexVectorTypeDef) String() string {
	return fmt.Sprintf("Vector[%s, %d]", td.ElemType.Name(), td.VectorLength)
}

type ComplexVectorView struct {
	SubtreeView
	*ComplexVectorTypeDef
}

func (tv *ComplexVectorView) HashTreeRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *ComplexVectorView) Get(i uint64) (View, error) {
	if i >= tv.ComplexVectorTypeDef.VectorLength {
		return nil, fmt.Errorf("cannot get item at element index %d, vector only has %d elements", i, tv.ComplexVectorTypeDef.VectorLength)
	}
	v, err := tv.SubtreeView.GetNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ComplexVectorTypeDef.ElemType.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ComplexVectorView) Set(i uint64, v View) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ComplexVectorView) setNode(i uint64, b Node) error {
	if i >= tv.ComplexVectorTypeDef.VectorLength {
		return fmt.Errorf("cannot set item at element index %d, vector only has %d elements", i, tv.ComplexVectorTypeDef.VectorLength)
	}
	return tv.SubtreeView.SetNode(i, b)
}

func (tv *ComplexVectorView) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}

func (tv *ComplexVectorView) Copy() (View, error) {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy, nil
}

func (tv *ComplexVectorView) ValueByteLength() (uint64, error) {
	if tv.IsFixedSize {
		return tv.Size, nil
	}
	// TODO
	return 0, nil
}

func (tv *ComplexVectorView) Serialize(w io.Writer) error {
	// TODO
	return nil
}

