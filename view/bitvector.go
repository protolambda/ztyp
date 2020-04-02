package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type BitVectorTypeDef struct {
	BitLength uint64
	ComplexTypeBase
}

func BitvectorType(name string, length uint64) *BitVectorTypeDef {
	byteSize := (length + 7) / 8
	return &BitVectorTypeDef{
		BitLength: length,
		ComplexTypeBase: ComplexTypeBase{
			TypeName:    name,
			MinSize:     byteSize,
			MaxSize:     byteSize,
			Size:        byteSize,
			IsFixedSize: true,
		},
	}
}

func (td *BitVectorTypeDef) Length() uint64 {
	return td.BitLength
}

func (td *BitVectorTypeDef) DefaultNode() Node {
	depth := CoverDepth(td.BottomNodeLength())
	return SubtreeFillToDepth(&ZeroHashes[0], depth)
}

func (td *BitVectorTypeDef) ViewFromBacking(node Node, hook BackingHook) (View, error) {
	depth := CoverDepth(td.BottomNodeLength())
	return &BitVectorView{
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
		BitVectorTypeDef: td,
	}, nil
}

func (td *BitVectorTypeDef) BottomNodeLength() uint64 {
	return (td.BitLength + 0xff) >> 8
}

func (td *BitVectorTypeDef) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *BitVectorTypeDef) New(hook BackingHook) *BitVectorView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*BitVectorView)
}

func (td *BitVectorTypeDef) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *BitVectorTypeDef) String() string {
	return fmt.Sprintf("Bitvector[%d]", td.BitLength)
}

type BitVectorView struct {
	SubtreeView
	*BitVectorTypeDef
}

func (tv *BitVectorView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = i>>8, uint8(i)
	v, err := tv.SubtreeView.GetNode(bottomIndex)
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
	return tv.SubtreeView.SetNode(bottomIndex, v.BackingFromBitfieldBase(r, subIndex))
}

// Shifts the bitvector contents to the right, clipping off the overflow. Only supported for small BitVectors.
func (tv *BitVectorView) ShRight(sh uint8) error {
	if tv.BitLength > 8 {
		return fmt.Errorf("shifting large bitvectors is unsupported")
	}
	v, err := tv.SubtreeView.GetNode(0)
	if err != nil {
		return err
	}
	r, ok := v.(*Root)
	if !ok {
		return fmt.Errorf("bitvector bottom node is not a root, cannot perform bitshift")
	}
	newRoot := *r
	// Mask to clip off bits.
	newRoot[0] = (newRoot[0] << sh) & ((1 << tv.BitLength) - 1)
	return tv.SubtreeView.SetNode(0, &newRoot)
}

func (tv *BitVectorView) Copy() (View, error) {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy, nil
}

func (tv *BitVectorView) Iter() BitIter {
	i := uint64(0)
	return BitIterFn(func() (elem bool, ok bool, err error) {
		if i < tv.BitLength {
			var item BoolView
			item, err = tv.Get(i)
			elem = bool(item)
			ok = true
			i += 1
			return
		} else {
			return false, false, nil
		}
	})
}

func (tv *BitVectorView) ReadonlyIter() BitIter {
	return bitReadonlyIter(tv.BackingNode, tv.BitLength, tv.depth)
}

func (tv *BitVectorView) ValueByteLength() (uint64, error) {
	return tv.Size, nil
}

func (tv *BitVectorView) Serialize(w io.Writer) error {
	// TODO
	return nil
}