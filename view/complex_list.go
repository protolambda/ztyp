package view

import (
	"encoding/binary"
	"fmt"

	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

type ComplexListTypeDef[EV View, ET TypeDef] struct {
	ElemType  ET
	ListLimit uint64
	ComplexTypeBase
}

var _ TypeDef[*ComplexListView[View, TypeDef]] = (*ComplexListTypeDef[View, TypeDef])(nil)

func ComplexListType[EV View, ET TypeDef](elemType ET, limit uint64) *ComplexListTypeDef[EV, ET] {
	maxSize := uint64(0)
	if elemType.IsFixedByteLength() {
		maxSize = limit * elemType.TypeByteLength()
	} else {
		maxSize = limit * (elemType.MaxByteLength() + OffsetByteLength)
	}
	return &ComplexListTypeDef[EV, ET]{
		ElemType:  elemType,
		ListLimit: limit,
		ComplexTypeBase: ComplexTypeBase{
			MinSize:     0,
			MaxSize:     maxSize,
			Size:        0,
			IsFixedSize: false,
		},
	}
}

func (td *ComplexListTypeDef[EV, ET]) ElementType() ET {
	return td.ElemType
}

func (td *ComplexListTypeDef[EV, ET]) Limit() uint64 {
	return td.ListLimit
}

func (td *ComplexListTypeDef[EV, ET]) DefaultNode() Node {
	depth := CoverDepth(td.ListLimit)
	// zeroed tree with zero mix-in
	return &PairNode{LeftChild: &ZeroHashes[depth], RightChild: &ZeroHashes[0]}
}

func (td *ComplexListTypeDef[EV, ET]) ViewFromBacking(node Node, hook BackingHook) (*ComplexListView[EV, ET], error) {
	depth := CoverDepth(td.ListLimit)
	return &ComplexListView[EV, ET]{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				Hook:        hook,
				BackingNode: node,
			},
			depth: depth + 1, // +1 for length mix-in
		},
		ComplexListTypeDef: td,
	}, nil
}

func (td *ComplexListTypeDef[EV, ET]) String() string {
	return fmt.Sprintf("List[%s, %d]", td.ElemType.String(), td.ListLimit)
}

type ComplexListView[EV View, ET TypeDef] struct {
	SubtreeView
	*ComplexListTypeDef[EV, ET]
}

var _ View = (*ComplexListView[View, TypeDef])(nil)

func AsComplexList[EV View, ET TypeDef](v View, err error) (*ComplexListView[EV, ET], error) {
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ComplexListView[EV, ET])
	if !ok {
		return nil, fmt.Errorf("view is not a list: %v", v)
	}
	return c, nil
}

func (td *ComplexListView[EV, ET]) Deserialize(dr *codec.DecodingReader) error {
	scope := dr.Scope()
	if scope == 0 {
		return td.SetBacking(td.DefaultNode())
	}
	if td.ElemType.IsFixedByteLength() {
		elemSize := td.ElemType.TypeByteLength()
		length := scope / elemSize
		if length > td.ListLimit {
			return fmt.Errorf("too many items, limit %d but got %d", td.ListLimit, length)
		}
		if expected := length * elemSize; expected != scope {
			return fmt.Errorf("scope %d does not align to elem size %d", scope, elemSize)
		}
		elements := make([]EV, length, length)
		for i := uint64(0); i < length; i++ {
			sub, err := dr.SubScope(elemSize)
			if err != nil {
				return err
			}
			el := td.ElemType.New()
			if err := el.Deserialize(sub); err != nil {
				return err
			}
			elements[i] = el
		}
		return td.SetElements(elements...)
	} else {
		firstOffset, err := dr.ReadOffset()
		if err != nil {
			return err
		}
		if firstOffset%OffsetByteLength != 0 {
			return fmt.Errorf("first offset %d does not align to offset length %d", firstOffset, OffsetByteLength)
		}
		length := uint64(firstOffset) / OffsetByteLength
		if length > td.ListLimit {
			return fmt.Errorf("too many items, limit %d but got %d", td.ListLimit, length)
		}
		offsets := make([]uint32, length, length)
		offsets[0] = firstOffset
		prevOffset := firstOffset
		for i := uint64(1); i < length; i++ {
			offset, err := dr.ReadOffset()
			if err != nil {
				return err
			}
			if offset < prevOffset {
				return fmt.Errorf("offset %d for element %d is smaller than previous offset %d", offset, i, prevOffset)
			}
			offsets[i] = offset
			prevOffset = offset
		}
		elements := make([]View, length, length)
		lastIndex := uint32(len(elements) - 1)
		for i := uint32(0); i < lastIndex; i++ {
			size := offsets[i+1] - offsets[i]
			sub, err := dr.SubScope(uint64(size))
			if err != nil {
				return err
			}
			el, err := td.ElemType.Deserialize(sub)
			if err != nil {
				return err
			}
			elements[i] = el
		}
		sub, err := dr.SubScope(scope - uint64(offsets[lastIndex]))
		if err != nil {
			return err
		}
		el, err := td.ElemType.Deserialize(sub)
		if err != nil {
			return err
		}
		elements[lastIndex] = el
		return td.FromElements(elements...)
	}
}

func (tv *ComplexListView[EV, ET]) SetElements(v ...EV) error {
	if uint64(len(v)) > tv.ListLimit {
		return fmt.Errorf("expected no more than %d elements, got %d", tv.ListLimit, len(v))
	}
	nodes := make([]Node, len(v), len(v))
	for i, el := range v {
		nodes[i] = el.Backing()
	}
	depth := CoverDepth(tv.ListLimit)
	contentsRootNode, _ := SubtreeFillToContents(nodes, depth)
	rootNode := &PairNode{LeftChild: contentsRootNode, RightChild: Uint64View(len(v)).Backing()}
	return tv.SetBacking(rootNode)
}

func (tv *ComplexListView[EV, ET]) Append(v View) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll >= tv.ListLimit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, tv.ListLimit)
	}
	// Appending is done by setting the node at the index list_length. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64(ll, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.BackingNode.Setter(lastGindex, true)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item: %v", err)
	}
	// Append the item by setting the newly allocated last item to it.
	// Update the view to the new tree containing this item.
	bNode, err := setLast(v.Backing())
	if err != nil {
		return err
	}
	// And update the list length
	setLength, err := bNode.Setter(RightGindex, false)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll+1)
	bNode, err = setLength(newLength)
	if err != nil {
		return err
	}
	return tv.SetBacking(bNode)
}

func (tv *ComplexListView[EV, ET]) Pop() error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no item can be popped")
	}
	// Popping is done by setting the node at the index list_length - 1. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64(ll, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.BackingNode.Setter(lastGindex, true)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item: %v", err)
	}
	// Pop the item by setting it to the zero hash
	// Update the view to the new tree containing this item.
	bNode, err := setLast(&ZeroHashes[0])
	// And update the list length
	setLength, err := bNode.Setter(RightGindex, false)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll-1)
	bNode, err = setLength(newLength)
	if err != nil {
		return err
	}
	return tv.SetBacking(bNode)
}

func (tv *ComplexListView[EV, ET]) CheckIndex(i uint64) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot handle item at element index %d, list only has %d elements", i, ll)
	}
	if i >= tv.ListLimit {
		return fmt.Errorf("list has a an invalid length of %d and cannot handle an element at index %d because of a limit of %d elements", ll, i, tv.ListLimit)
	}
	return nil
}

func (tv *ComplexListView[EV, ET]) Get(i uint64) (EV, error) {
	var out EV
	if err := tv.CheckIndex(i); err != nil {
		return out, err
	}
	v, err := tv.SubtreeView.GetNode(i)
	if err != nil {
		return out, err
	}
	return tv.ComplexListTypeDef.ElemType.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ComplexListView[EV, ET]) Set(i uint64, v View) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ComplexListView[EV, ET]) setNode(i uint64, b Node) error {
	if err := tv.CheckIndex(i); err != nil {
		return err
	}
	return tv.SubtreeView.SetNode(i, b)
}

func (tv *ComplexListView[EV, ET]) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}

func (tv *ComplexListView[EV, ET]) Length() (uint64, error) {
	v, err := tv.SubtreeView.BackingNode.Getter(RightGindex)
	if err != nil {
		return 0, err
	}
	llBytes, ok := v.(*Root)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	ll := binary.LittleEndian.Uint64(llBytes[:8])
	if ll > tv.ListLimit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return ll, nil
}

func (tv *ComplexListView[EV, ET]) Copy() *ComplexListView[EV, ET] {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy
}

func (tv *ComplexListView[EV, ET]) Iter() ElemIter[EV, ET] {
	length, err := tv.Length()
	if err != nil {
		return ErrElemIter[EV, ET]{err}
	}
	i := uint64(0)
	return ElemIterFn[EV, ET](func() (elem EV, elemType ET, ok bool, err error) {
		if i < length {
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

func (tv *ComplexListView[EV, ET]) ReadonlyIter() ElemIter[EV, ET] {
	length, err := tv.Length()
	if err != nil {
		return ErrElemIter[EV, ET]{err}
	}
	// get contents subtree, to traverse with the stack
	node, err := tv.BackingNode.Left()
	if err != nil {
		return ErrElemIter[EV, ET]{err}
	}
	// ignore length mixin in stack
	return elemReadonlyIter[EV, ET](node, length, tv.depth-1, tv.ElemType)
}

func (tv *ComplexListView[EV, ET]) ValueByteLength() (uint64, error) {
	length, err := tv.Length()
	if err != nil {
		return 0, err
	}
	if tv.ElemType.IsFixedByteLength() {
		return length * tv.ElemType.TypeByteLength(), nil
	} else {
		size := length * OffsetByteLength
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

func (tv *ComplexListView[EV, ET]) Serialize(w *codec.EncodingWriter) error {
	if tv.ElemType.IsFixedByteLength() {
		return serializeComplexFixElemSeries(tv.ReadonlyIter(), w)
	} else {
		length, err := tv.Length()
		if err != nil {
			return err
		}
		return serializeComplexVarElemSeries(length, tv.ReadonlyIter, w)
	}
}
