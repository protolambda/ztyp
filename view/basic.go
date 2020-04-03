package view

import (
	"encoding/binary"
	"errors"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

// A uint type, identified by its size in bytes.
type UintMeta uint64

func (td UintMeta) Default(_ BackingHook) View {
	switch td {
	case Uint8Type:
		return Uint8View(0)
	case Uint16Type:
		return Uint16View(0)
	case Uint32Type:
		return Uint32View(0)
	case Uint64Type:
		return Uint64View(0)
	default:
		// unsupported uint type
		return nil
	}
}

func (td UintMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (td UintMeta) New() BasicView {
	switch td {
	case Uint8Type:
		return Uint8View(0)
	case Uint16Type:
		return Uint16View(0)
	case Uint32Type:
		return Uint32View(0)
	case Uint64Type:
		return Uint64View(0)
	default:
		return nil
	}
}

var UnsupportedUintType = errors.New("unsupported uint type")

func (td UintMeta) ViewFromBacking(node Node, _ BackingHook) (View, error) {
	v, ok := node.(*Root)
	if !ok {
		return nil, fmt.Errorf("node %v must be a root to read a uint%d from it", node, td)
	}
	switch td {
	case Uint8Type:
		return Uint8View(v[0]), nil
	case Uint16Type:
		return Uint16View(binary.LittleEndian.Uint16(v[:2])), nil
	case Uint32Type:
		return Uint32View(binary.LittleEndian.Uint32(v[:4])), nil
	case Uint64Type:
		return Uint64View(binary.LittleEndian.Uint64(v[:8])), nil
	default:
		return nil, UnsupportedUintType
	}
}

func (td UintMeta) BasicViewFromBacking(v *Root, i uint8) (BasicView, error) {
	if uint64(i) >= (32 / uint64(td)) {
		return nil, fmt.Errorf("cannot get subview %d for uint%d type", i, td)
	}
	switch td {
	case Uint8Type:
		return Uint8View(v[i]), nil
	case Uint16Type:
		return Uint16View(binary.LittleEndian.Uint16(v[2*i : 2*i+2])), nil
	case Uint32Type:
		return Uint32View(binary.LittleEndian.Uint32(v[4*i : 4*i+4])), nil
	case Uint64Type:
		return Uint64View(binary.LittleEndian.Uint64(v[8*i : 8*i+8])), nil
	default:
		return nil, UnsupportedUintType
	}
}

func (td UintMeta) PackViews(views []BasicView) ([]Node, error) {
	// TODO can be optimized: switch on type, put contents directly into node bytes, and don't use temporary nodes
	perNode := uint8(32 / td)
	chunkCount := (uint64(len(views)) + uint64(perNode) - 1) / uint64(perNode)
	chunks := make([]Node, chunkCount, chunkCount)
	i := 0
	for chunk := uint64(0); chunk < chunkCount; chunk++ {
		root := &Root{}
		for j := uint8(0); j < perNode && i < len(views); j++ {
			root = views[i].BackingFromBase(root, j)
			i += 1
		}
		chunks = append(chunks, root)
	}
	return chunks, nil
}

func (td UintMeta) IsFixedByteLength() bool {
	return true
}

func (td UintMeta) TypeByteLength() uint64 {
	return uint64(td)
}

func (td UintMeta) MinByteLength() uint64 {
	return uint64(td)
}

func (td UintMeta) MaxByteLength() uint64 {
	return uint64(td)
}

func (td UintMeta) Deserialize(r io.Reader, scope uint64) (View, error) {
	if scope < uint64(td) {
		return nil, fmt.Errorf("scope of %d bytes not enough for %s", scope, td.String())
	}
	tmp := make([]byte, td, td)
	_, err := r.Read(tmp)
	if err != nil {
		return nil, err
	}
	switch td {
	case Uint8Type:
		return Uint8View(tmp[0]), nil
	case Uint16Type:
		return Uint16View(binary.LittleEndian.Uint16(tmp)), nil
	case Uint32Type:
		return Uint32View(binary.LittleEndian.Uint32(tmp)), nil
	case Uint64Type:
		return Uint64View(binary.LittleEndian.Uint64(tmp)), nil
	default:
		return nil, UnsupportedUintType
	}
}

func (td UintMeta) String() string {
	return fmt.Sprintf("uint%d", td*8)
}

const (
	Uint8Type  UintMeta = 1
	Uint16Type UintMeta = 2
	Uint32Type UintMeta = 4
	Uint64Type UintMeta = 8
)

var BasicViewNoSetBackingError = errors.New("basic views cannot set new backing")

type Uint8View uint8

func (v Uint8View) SetBacking(b Node) error {
	return BasicViewNoSetBackingError
}

func (v Uint8View) Backing() Node {
	out := &Root{}
	out[0] = uint8(v)
	return out
}

func (v Uint8View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 32 {
		return nil
	}
	newRoot := *base
	newRoot[i] = uint8(v)
	return &newRoot
}

func (v Uint8View) Copy() (View, error) {
	return v, nil
}

func (v Uint8View) ValueByteLength() (uint64, error) {
	return 1, nil
}

func (v Uint8View) Serialize(w io.Writer) error {
	_, err := w.Write([]byte{byte(v)})
	return err
}

func (v Uint8View) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	newRoot[0] = uint8(v)
	return newRoot
}

func (v Uint8View) Type() TypeDef {
	return Uint8Type
}

// Alias to Uint8Type
const ByteType = Uint8Type

// Alias to Uint8View
type ByteView = Uint8View

type Uint16View uint16

func (v Uint16View) SetBacking(b Node) error {
	return BasicViewNoSetBackingError
}

func (v Uint16View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint16(out[:2], uint16(v))
	return out
}

func (v Uint16View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 16 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint16(newRoot[i<<1:(i<<1)+2], uint16(v))
	return &newRoot
}

func (v Uint16View) Copy() (View, error) {
	return v, nil
}

func (v Uint16View) ValueByteLength() (uint64, error) {
	return 2, nil
}

func (v Uint16View) Serialize(w io.Writer) error {
	_, err := w.Write([]byte{byte(v), byte(v >> 8)})
	return err
}

func (v Uint16View) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	binary.LittleEndian.PutUint16(newRoot[:], uint16(v))
	return newRoot
}

func (v Uint16View) Type() TypeDef {
	return Uint16Type
}

type Uint32View uint32

func (v Uint32View) SetBacking(b Node) error {
	return BasicViewNoSetBackingError
}

func (v Uint32View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint32(out[:4], uint32(v))
	return out
}

func (v Uint32View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 8 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint32(newRoot[i*4:i*4+4], uint32(v))
	return &newRoot
}

func (v Uint32View) Copy() (View, error) {
	return v, nil
}

func (v Uint32View) ValueByteLength() (uint64, error) {
	return 4, nil
}

func (v Uint32View) Serialize(w io.Writer) error {
	_, err := w.Write([]byte{byte(v), byte(v >> 8), byte(v >> 16), byte(v >> 24)})
	return err
}

func (v Uint32View) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	binary.LittleEndian.PutUint32(newRoot[:], uint32(v))
	return newRoot
}

func (v Uint32View) Type() TypeDef {
	return Uint32Type
}

type Uint64View uint64

func (v Uint64View) SetBacking(b Node) error {
	return BasicViewNoSetBackingError
}

func (v Uint64View) Backing() Node {
	out := &Root{}
	binary.LittleEndian.PutUint64(out[:8], uint64(v))
	return out
}

func (v Uint64View) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 4 {
		return nil
	}
	newRoot := *base
	binary.LittleEndian.PutUint64(newRoot[i*8:i*8+8], uint64(v))
	return &newRoot
}

func (v Uint64View) Copy() (View, error) {
	return v, nil
}

func (v Uint64View) ValueByteLength() (uint64, error) {
	return 8, nil
}

func (v Uint64View) Serialize(w io.Writer) error {
	b := [8]byte{}
	binary.LittleEndian.PutUint64(b[:], uint64(v))
	_, err := w.Write(b[:])
	return err
}

func (v Uint64View) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	binary.LittleEndian.PutUint64(newRoot[:], uint64(v))
	return newRoot
}

func (v Uint64View) Type() TypeDef {
	return Uint64Type
}

type BoolMeta uint8

func (td BoolMeta) SubViewFromBacking(v *Root, i uint8) BasicView {
	if i >= 32 {
		return nil
	}
	if v[i] > 1 {
		return nil
	}
	return BoolView(v[i] == 1)
}

func (td BoolMeta) BoolViewFromBitfieldBacking(v *Root, i uint8) (BoolView, error) {
	if i > 32 {
		return false, fmt.Errorf("out of range bit lookup in node: index: %d root: %x", i, v)
	}
	return (v[i>>3]>>(i&7))&1 == 1, nil
}

func (td BoolMeta) Default(_ BackingHook) View {
	return BoolView(false)
}

func (td BoolMeta) New() BoolView {
	return false
}

func (td BoolMeta) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (td BoolMeta) ViewFromBacking(node Node, _ BackingHook) (View, error) {
	v, ok := node.(*Root)
	if !ok {
		return nil, fmt.Errorf("node %v must be a root to read a bool from it", node)
	}
	return BoolView(v[0] != 0), nil
}

func (td BoolMeta) IsFixedByteLength() bool {
	return true
}

func (td BoolMeta) TypeByteLength() uint64 {
	return 1
}

func (td BoolMeta) MinByteLength() uint64 {
	return 1
}

func (td BoolMeta) MaxByteLength() uint64 {
	return 1
}

func (td BoolMeta) Deserialize(r io.Reader, scope uint64) (View, error) {
	if scope == 0 {
		return nil, errors.New("cannot read bool, scope is 0")
	}
	b := [1]byte{}
	_, err := r.Read(b[:])
	if err != nil {
		return nil, err
	}
	if b[0] > 1 {
		return nil, fmt.Errorf("invalid bool value: 0x%x", b[0])
	}
	return BoolView(b[0] == 1), nil
}

func (td BoolMeta) String() string {
	return "bool"
}

const BoolType BoolMeta = 0

type BoolView bool

var trueRoot = &Root{1}

func (v BoolView) SetBacking(b Node) error {
	return BasicViewNoSetBackingError
}

func (v BoolView) Backing() Node {
	if v {
		return trueRoot
	} else {
		return &ZeroHashes[0]
	}
}

func (v BoolView) BackingFromBitfieldBase(base *Root, i uint8) *Root {
	newRoot := *base
	if v {
		newRoot[i>>3] |= 1 << (i & 7)
	} else {
		newRoot[i>>3] &^= 1 << (i & 7)
	}
	return &newRoot
}

func (v BoolView) byte() byte {
	if v {
		return 1
	} else {
		return 0
	}
}

func (v BoolView) BackingFromBase(base *Root, i uint8) *Root {
	if i >= 32 {
		return nil
	}
	newRoot := *base
	if v {
		newRoot[i] = 1
	} else {
		newRoot[i] = 0
	}
	return &newRoot
}

func (v BoolView) Copy() (View, error) {
	return v, nil
}

func (v BoolView) ValueByteLength() (uint64, error) {
	return 1, nil
}

func (v BoolView) Serialize(w io.Writer) error {
	_, err := w.Write([]byte{v.byte()})
	return err
}

func (v BoolView) HashTreeRoot(h HashFn) Root {
	newRoot := Root{}
	newRoot[0] = v.byte()
	return newRoot
}

func (v BoolView) Type() TypeDef {
	return BoolType
}
