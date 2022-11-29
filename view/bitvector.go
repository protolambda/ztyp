package view

import (
	"fmt"

	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

type BitVectorTypeDef struct {
	BitLength uint64
	ComplexTypeBase
}

func BitVectorType(length uint64) *BitVectorTypeDef {
	byteSize := (length + 7) / 8
	return &BitVectorTypeDef{
		BitLength: length,
		ComplexTypeBase: ComplexTypeBase{
			MinSize:     byteSize,
			MaxSize:     byteSize,
			Size:        byteSize,
			IsFixedSize: true,
		},
	}
}

func (td *BitVectorTypeDef) New() MutView {
	depth := CoverDepth(td.BottomNodeLength())
	return &BitVectorView{
		SubtreeView: SubtreeView{
			BackedView: BackedView{},
			depth:      depth,
		},
		BitVectorTypeDef: td,
	}
}

func (td *BitVectorTypeDef) Length() uint64 {
	return td.BitLength
}

func (td *BitVectorTypeDef) DefaultNode() Node {
	depth := CoverDepth(td.BottomNodeLength())
	return SubtreeFillToDepth(&ZeroHashes[0], depth)
}

func (td *BitVectorTypeDef) BottomNodeLength() uint64 {
	return (td.BitLength + 0xff) >> 8
}

func (td *BitVectorTypeDef) String() string {
	return fmt.Sprintf("Bitvector[%d]", td.BitLength)
}

type BitVectorView struct {
	SubtreeView
	*BitVectorTypeDef
}

func (td *BitVectorView) SetBits(bits []bool) error {
	if uint64(len(bits)) != td.BitLength {
		return fmt.Errorf("got %d bits, expected %d bits", len(bits), td.BitLength)
	}
	contents := bitsToBytes(bits)
	bottomNodes, err := BytesIntoNodes(contents)
	if err != nil {
		return err
	}
	depth := CoverDepth(td.BottomNodeLength())
	rootNode, _ := SubtreeFillToContents(bottomNodes, depth)
	return td.SetBacking(rootNode)
}

func (td *BitVectorView) Deserialize(dr *codec.DecodingReader) error {
	scope := dr.Scope()
	if td.Size != scope {
		return fmt.Errorf("expected size %d does not match scope %d", td.Size, scope)
	}
	contents := make([]byte, scope, scope)
	if _, err := dr.Read(contents); err != nil {
		return err
	}
	if scope != 0 && td.BitLength&7 != 0 {
		last := contents[scope-1]
		if last&byte((uint16(1)<<(td.BitLength&7))-1) != last {
			return fmt.Errorf("last bitvector byte %d has out of bounds bits set", last)
		}
	}
	bottomNodes, err := BytesIntoNodes(contents)
	if err != nil {
		return err
	}
	depth := CoverDepth(td.BottomNodeLength())
	rootNode, _ := SubtreeFillToContents(bottomNodes, depth)
	return td.SetBacking(rootNode)
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
	return BoolType{}.BoolViewFromBitfieldBacking(r, subIndex)
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

func (tv *BitVectorView) Copy() *BitVectorView {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy
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

func (tv *BitVectorView) Serialize(w *codec.EncodingWriter) error {
	contents := make([]byte, tv.Size, tv.Size)
	if err := SubtreeIntoBytes(tv.BackingNode, tv.depth, tv.BottomNodeLength(), contents); err != nil {
		return err
	}
	return w.Write(contents)
}
