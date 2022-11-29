package view

import (
	"fmt"

	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

type ComplexVectorTypeDef[EV View, ET TypeDef[EV]] struct {
	ElemType     ET
	VectorLength uint64
	ComplexTypeBase
}

var _ TypeDef[*ComplexVectorView[View, TypeDef[View]]] = (*ComplexVectorTypeDef[View, TypeDef[View]])(nil)

func ComplexVectorType[EV View, ET TypeDef[EV]](elemType ET, length uint64) *ComplexVectorTypeDef[EV, ET] {
	minSize := uint64(0)
	maxSize := uint64(0)
	size := uint64(0)
	isFixedSize := elemType.IsFixedByteLength()
	if isFixedSize {
		size = length * elemType.TypeByteLength()
		minSize = size
		maxSize = size
	} else {
		minSize = length * (elemType.MinByteLength() + OffsetByteLength)
		maxSize = length * (elemType.MaxByteLength() + OffsetByteLength)
	}
	return &ComplexVectorTypeDef[EV, ET]{
		ElemType:     elemType,
		VectorLength: length,
		ComplexTypeBase: ComplexTypeBase{
			MinSize:     minSize,
			MaxSize:     maxSize,
			Size:        size,
			IsFixedSize: isFixedSize,
		},
	}
}

func (td *ComplexVectorTypeDef[EV, ET]) FromElements(v ...View) (*ComplexVectorView[EV, ET], error) {
	if td.VectorLength != uint64(len(v)) {
		return nil, fmt.Errorf("expected %d elements, got %d", td.VectorLength, len(v))
	}
	nodes := make([]Node, td.VectorLength, td.VectorLength)
	for i, el := range v {
		nodes[i] = el.Backing()
	}
	depth := CoverDepth(td.VectorLength)
	rootNode, _ := SubtreeFillToContents(nodes, depth)
	vecView, _ := td.ViewFromBacking(rootNode, nil)
	return vecView, nil
}

func (td *ComplexVectorTypeDef[EV, ET]) ElementType() ET {
	return td.ElemType
}

func (td *ComplexVectorTypeDef[EV, ET]) Length() uint64 {
	return td.VectorLength
}

func (td *ComplexVectorTypeDef[EV, ET]) DefaultNode() Node {
	depth := CoverDepth(td.VectorLength)
	// The same node N times: the node is immutable, so re-use is safe.
	defaultNode := td.ElemType.DefaultNode()
	// can ignore error, depth is derived from length.
	rootNode, _ := SubtreeFillToLength(defaultNode, depth, td.VectorLength)
	return rootNode
}

func (td *ComplexVectorTypeDef[EV, ET]) ViewFromBacking(node Node, hook BackingHook) (*ComplexVectorView[EV, ET], error) {
	depth := CoverDepth(td.VectorLength)
	return &ComplexVectorView[EV, ET]{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				Hook:        hook,
				BackingNode: node,
			},
			depth: depth,
		},
		ComplexVectorTypeDef: td,
	}, nil
}

func (td *ComplexVectorTypeDef[EV, ET]) Deserialize(dr *codec.DecodingReader) (*ComplexVectorView[EV, ET], error) {
	scope := dr.Scope()
	if td.IsFixedSize {
		elemSize := td.ElemType.TypeByteLength()
		if td.Size != scope {
			return nil, fmt.Errorf("expected size %d does not match scope %d", td.Size, scope)
		}
		elements := make([]View, td.VectorLength, td.VectorLength)
		for i := uint64(0); i < td.VectorLength; i++ {
			sub, err := dr.SubScope(elemSize)
			if err != nil {
				return nil, err
			}
			el, err := td.ElemType.Deserialize(sub)
			if err != nil {
				return nil, err
			}
			elements[i] = el
		}
		return td.FromElements(elements...)
	} else {
		offsets := make([]uint32, td.VectorLength, td.VectorLength)
		prevOffset := uint32(0)
		for i := uint64(0); i < td.VectorLength; i++ {
			offset, err := dr.ReadOffset()
			if err != nil {
				return nil, err
			}
			if offset < prevOffset {
				return nil, fmt.Errorf("offset %d for element %d is smaller than previous offset %d", offset, i, prevOffset)
			}
			offsets[i] = offset
			prevOffset = offset
		}
		elements := make([]View, td.VectorLength, td.VectorLength)
		lastIndex := uint32(len(elements) - 1)
		for i := uint32(0); i < lastIndex; i++ {
			size := offsets[i+1] - offsets[i]
			sub, err := dr.SubScope(uint64(size))
			if err != nil {
				return nil, err
			}
			el, err := td.ElemType.Deserialize(sub)
			if err != nil {
				return nil, err
			}
			elements[i] = el
		}
		sub, err := dr.SubScope(scope - uint64(offsets[lastIndex]))
		if err != nil {
			return nil, err
		}
		el, err := td.ElemType.Deserialize(sub)
		if err != nil {
			return nil, err
		}
		elements[lastIndex] = el
		return td.FromElements(elements...)
	}
}

func (td *ComplexVectorTypeDef[EV, ET]) String() string {
	return fmt.Sprintf("Vector[%s, %d]", td.ElemType.String(), td.VectorLength)
}

type ComplexVectorView[EV View, ET TypeDef[EV]] struct {
	SubtreeView
	*ComplexVectorTypeDef[EV, ET]
}

var _ View = (*ComplexVectorView[View, TypeDef[View]])(nil)

func AsComplexVector[EV View, ET TypeDef[EV]](v View, err error) (*ComplexVectorView[EV, ET], error) {
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ComplexVectorView[EV, ET])
	if !ok {
		return nil, fmt.Errorf("view is not a vector: %v", v)
	}
	return c, nil
}

func (tv *ComplexVectorView[EV, ET]) Get(i uint64) (EV, error) {
	var out EV
	if i >= tv.ComplexVectorTypeDef.VectorLength {
		return out, fmt.Errorf("cannot get item at element index %d, vector only has %d elements", i, tv.ComplexVectorTypeDef.VectorLength)
	}
	v, err := tv.SubtreeView.GetNode(i)
	if err != nil {
		return out, err
	}
	return tv.ComplexVectorTypeDef.ElemType.ViewFromBacking(v, tv.ItemHook(i)), nil
}

func (tv *ComplexVectorView[EV, ET]) Set(i uint64, v EV) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ComplexVectorView[EV, ET]) setNode(i uint64, b Node) error {
	if i >= tv.ComplexVectorTypeDef.VectorLength {
		return fmt.Errorf("cannot set item at element index %d, vector only has %d elements", i, tv.ComplexVectorTypeDef.VectorLength)
	}
	return tv.SubtreeView.SetNode(i, b)
}

func (tv *ComplexVectorView[EV, ET]) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}

func (tv *ComplexVectorView[EV, ET]) Copy() *ComplexVectorView[EV, ET] {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy
}

func (tv *ComplexVectorView[EV, ET]) Iter() ElemIter[EV, ET] {
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

func (tv *ComplexVectorView[EV, ET]) ReadonlyIter() ElemIter[EV, ET] {
	return elemReadonlyIter[EV, ET](tv.BackingNode, tv.VectorLength, tv.depth, tv.ElemType)
}

func (tv *ComplexVectorView[EV, ET]) ValueByteLength() (uint64, error) {
	if tv.IsFixedSize {
		return tv.Size, nil
	} else {
		size := tv.VectorLength * OffsetByteLength
		iter := tv.ReadonlyIter()
		for {
			elem, _, ok, err := iter.Next()
			if err != nil {
				return 0, err
			}
			if !ok {
				break
			}
			valSize, err := elem.ValueByteLength()
			if err != nil {
				return 0, err
			}
			size += valSize
		}
		return size, nil
	}
}

func (tv *ComplexVectorView[EV, ET]) Serialize(w *codec.EncodingWriter) error {
	if tv.IsFixedSize {
		return serializeComplexFixElemSeries(tv.ReadonlyIter(), w)
	} else {
		return serializeComplexVarElemSeries(tv.VectorLength, tv.ReadonlyIter, w)
	}
}
