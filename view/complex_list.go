package view

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type ComplexListTypeDef struct {
	ElemType  TypeDef
	ListLimit uint64
	ComplexTypeBase
}

func ComplexListType(name string, elemType TypeDef, limit uint64) *ComplexListTypeDef {
	maxSize := uint64(0)
	if elemType.IsFixedByteLength() {
		maxSize = limit * elemType.TypeByteLength()
	} else {
		maxSize = limit * elemType.MaxByteLength()
	}
	return &ComplexListTypeDef{
		ElemType:  elemType,
		ListLimit: limit,
		ComplexTypeBase: ComplexTypeBase{
			TypeName:    name,
			MinSize:     0,
			MaxSize:     maxSize,
			Size:        0,
			IsFixedSize: false,
		},
	}
}

func (td *ComplexListTypeDef) ElementType() TypeDef {
	return td.ElemType
}

func (td *ComplexListTypeDef) Limit() uint64 {
	return td.ListLimit
}

func (td *ComplexListTypeDef) DefaultNode() Node {
	depth := CoverDepth(td.ListLimit)
	// zeroed tree with zero mix-in
	return &PairNode{LeftChild: &ZeroHashes[depth], RightChild: &ZeroHashes[0]}
}

func (td *ComplexListTypeDef) ViewFromBacking(node Node, hook BackingHook) (View, error) {
	depth := CoverDepth(td.ListLimit)
	return &ComplexListView{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				ViewBase: ViewBase{
					TypeDef: td,
				},
				Hook:        hook,
				BackingNode: node,
			},
			depth: depth + 1, // +1 for length mix-in
		},
		ComplexListTypeDef: td,
	}, nil
}

func (td *ComplexListTypeDef) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *ComplexListTypeDef) New(hook BackingHook) *ComplexListView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*ComplexListView)
}

func (td *ComplexListTypeDef) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *ComplexListTypeDef) String() string {
	return fmt.Sprintf("List[%s, %d]", td.ElemType.Name(), td.ListLimit)
}

type ComplexListView struct {
	SubtreeView
	*ComplexListTypeDef
}

func (tv *ComplexListView) Append(v View) error {
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

func (tv *ComplexListView) Pop() error {
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

func (tv *ComplexListView) CheckIndex(i uint64) error {
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

func (tv *ComplexListView) Get(i uint64) (View, error) {
	if err := tv.CheckIndex(i); err != nil {
		return nil, err
	}
	v, err := tv.SubtreeView.GetNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ComplexListTypeDef.ElemType.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ComplexListView) Set(i uint64, v View) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ComplexListView) setNode(i uint64, b Node) error {
	if err := tv.CheckIndex(i); err != nil {
		return err
	}
	return tv.SubtreeView.SetNode(i, b)
}

func (tv *ComplexListView) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}

func (tv *ComplexListView) Length() (uint64, error) {
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

func (tv *ComplexListView) Copy() (View, error) {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy, nil
}

func (tv *ComplexListView) Iter() ElemIter {
	length, err := tv.Length()
	if err != nil {
		return ErrIter{err}
	}
	i := uint64(0)
	return ElemIterFn(func() (elem View, ok bool, err error) {
		if i < length {
			elem, err = tv.Get(i)
			ok = true
			i += 1
			return
		} else {
			return nil, false, nil
		}
	})
}

func (tv *ComplexListView) ReadonlyIter() ElemIter {
	length, err := tv.Length()
	if err != nil {
		return ErrIter{err}
	}
	// get contents subtree, to traverse with the stack
	node, err := tv.BackingNode.Left()
	if err != nil {
		return ErrIter{err}
	}
	// ignore length mixin in stack
	return elemReadonlyIter(node, length, tv.depth - 1, tv.ElemType)
}

func (tv *ComplexListView) ValueByteLength() (uint64, error) {
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
			elem, ok, err := iter.Next()
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

func (tv *ComplexListView) Serialize(w io.Writer) error {
	// TODO
	return nil
}

