package view

import (
	"encoding/binary"
	"errors"
	"fmt"
	"strconv"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	. "github.com/protolambda/ztyp/tree"
)

type PackingType[V View] interface {
	TypeDef
	PackViews(views []V) ([]Node, error)
}

type UintView interface {
	Uint8View | Uint16View | Uint32View | Uint64View | Uint256View
	BasicView
}

type UintType interface {
	Uint8Type | Uint16Type | Uint32Type | Uint64Type | Uint256Type
	TypeDef
}

type Uint8Type struct{}

var _ PackingType[Uint8View] = Uint8Type{}

func (Uint8Type) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (Uint8Type) New() MutView {
	return new(Uint8View)
}

func (Uint8Type) PackViews(views []Uint8View) ([]Node, error) {
	out := make([]Node, (len(views)+31)/32)
	j := 0
	for i := range out {
		var x Root
		for h := 0; h < 32 && j < len(views); j++ {
			x[h] = byte(views[j])
			h++
		}
		out[i] = &x
	}
	return out, nil
}

func (Uint8Type) IsFixedByteLength() bool {
	return true
}

func (Uint8Type) TypeByteLength() uint64 {
	return 1
}

func (Uint8Type) MinByteLength() uint64 {
	return 1
}

func (Uint8Type) MaxByteLength() uint64 {
	return 1
}

func (Uint8Type) Deserialize(dr *codec.DecodingReader) (Uint8View, error) {
	v, err := dr.ReadByte()
	return Uint8View(v), err
}

func (td Uint8Type) String() string {
	return "uint8"
}

type Uint16Type struct{}

var _ PackingType[Uint16View] = Uint16Type{}

func (Uint16Type) Default(_ BackingHook) Uint16View {
	return 0
}

func (Uint16Type) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (Uint16Type) New() MutView {
	return new(Uint16View)
}

func (Uint16Type) ViewFromBacking(node Node, _ BackingHook) (Uint16View, error) {
	v, ok := node.(*Root)
	if !ok {
		return 0, fmt.Errorf("node %T must be a root to read a uint16 from it", node)
	}
	return Uint16View(binary.LittleEndian.Uint16(v[0:2])), nil
}

func (Uint16Type) BasicViewFromBacking(v *Root, i uint8) (Uint16View, error) {
	if i >= 16 {
		return 0, fmt.Errorf("cannot get uint16 at %d in 32 byte root", i)
	}
	return Uint16View(binary.LittleEndian.Uint16(v[i<<1:])), nil
}

func (Uint16Type) PackViews(views []Uint16View) ([]Node, error) {
	out := make([]Node, (len(views)+15)/16)
	j := 0
	for i := range out {
		var x Root
		for h := 0; h < 32 && j < len(views); j++ {
			x[h] = byte(views[j])
			h++
			x[h] = byte(views[j] >> 8)
			h++
		}
		out[i] = &x
	}
	return out, nil
}

func (Uint16Type) IsFixedByteLength() bool {
	return true
}

func (Uint16Type) TypeByteLength() uint64 {
	return 2
}

func (Uint16Type) MinByteLength() uint64 {
	return 2
}

func (Uint16Type) MaxByteLength() uint64 {
	return 2
}

func (Uint16Type) Deserialize(dr *codec.DecodingReader) (Uint16View, error) {
	v, err := dr.ReadUint16()
	return Uint16View(v), err
}

func (td Uint16Type) String() string {
	return "uint16"
}

type Uint32Type struct{}

var _ PackingType[Uint32View] = Uint32Type{}

func (Uint32Type) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (Uint32Type) New() MutView {
	return new(Uint32View)
}

func (Uint32Type) ViewFromBacking(node Node, _ BackingHook) (Uint32View, error) {
	v, ok := node.(*Root)
	if !ok {
		return 0, fmt.Errorf("node %T must be a root to read a uint32 from it", node)
	}
	return Uint32View(binary.LittleEndian.Uint32(v[0:4])), nil
}

func (Uint32Type) BasicViewFromBacking(v *Root, i uint8) (Uint32View, error) {
	if i >= 8 {
		return 0, fmt.Errorf("cannot get uint32 at %d in 32 byte root", i)
	}
	return Uint32View(binary.LittleEndian.Uint32(v[i<<2:])), nil
}

func (Uint32Type) PackViews(views []Uint32View) ([]Node, error) {
	out := make([]Node, (len(views)+7)/8)
	j := 0
	for i := range out {
		var x Root
		for h := 0; h < 32 && j < len(views); j++ {
			y := views[j]
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
		}
		out[i] = &x
	}
	return out, nil
}

func (Uint32Type) IsFixedByteLength() bool {
	return true
}

func (Uint32Type) TypeByteLength() uint64 {
	return 4
}

func (Uint32Type) MinByteLength() uint64 {
	return 4
}

func (Uint32Type) MaxByteLength() uint64 {
	return 4
}

func (Uint32Type) Deserialize(dr *codec.DecodingReader) (Uint32View, error) {
	v, err := dr.ReadUint32()
	return Uint32View(v), err
}

func (td Uint32Type) String() string {
	return "uint32"
}

type Uint64Type struct{}

var _ PackingType[Uint64View] = Uint64Type{}

func (Uint64Type) Default(_ BackingHook) Uint64View {
	return 0
}

func (Uint64Type) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (Uint64Type) New() MutView {
	return new(Uint64View)
}

func (Uint64Type) ViewFromBacking(node Node, _ BackingHook) (Uint64View, error) {
	v, ok := node.(*Root)
	if !ok {
		return 0, fmt.Errorf("node %T must be a root to read a uint64 from it", node)
	}
	return Uint64View(binary.LittleEndian.Uint64(v[0:8])), nil
}

func (Uint64Type) BasicViewFromBacking(v *Root, i uint8) (Uint64View, error) {
	if i >= 4 {
		return 0, fmt.Errorf("cannot get uint64 at %d in 32 byte root", i)
	}
	return Uint64View(binary.LittleEndian.Uint64(v[i<<3:])), nil
}

func (Uint64Type) PackViews(views []Uint64View) ([]Node, error) {
	out := make([]Node, (len(views)+3)/4)
	j := 0
	for i := range out {
		var x Root
		for h := 0; h < 32 && j < len(views); j++ {
			y := views[j]
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
			y >>= 8
			x[h] = byte(y)
			h++
		}
		out[i] = &x
	}
	return out, nil
}

func (Uint64Type) IsFixedByteLength() bool {
	return true
}

func (Uint64Type) TypeByteLength() uint64 {
	return 8
}

func (Uint64Type) MinByteLength() uint64 {
	return 8
}

func (Uint64Type) MaxByteLength() uint64 {
	return 8
}

func (Uint64Type) Deserialize(dr *codec.DecodingReader) (Uint64View, error) {
	v, err := dr.ReadUint64()
	return Uint64View(v), err
}

func (td Uint64Type) String() string {
	return "uint64"
}

var BasicBackingRequiredErr = errors.New("basic views require a basic Root backing")

var BadLengthError = errors.New("scope is wrong")

type Uint8View uint8

var _ BasicView = (*Uint8View)(nil)

func (v *Uint8View) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		*v = Uint8View(r[0])
		return nil
	} else {
		return BasicBackingRequiredErr
	}
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

func (v Uint8View) Copy() (Uint8View, error) {
	return v, nil
}

func (v Uint8View) ValueByteLength() (uint64, error) {
	return 1, nil
}

func (v Uint8View) ByteLength() uint64 {
	return 1
}

func (v Uint8View) FixedLength() uint64 {
	return 1
}

func (v Uint8View) Serialize(w *codec.EncodingWriter) error {
	return w.WriteByte(byte(v))
}

func (v Uint8View) Encode() ([]byte, error) {
	return []byte{byte(v)}, nil
}

func (v *Uint8View) Deserialize(r *codec.DecodingReader) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	*v = Uint8View(b)
	return nil
}

func (v *Uint8View) Decode(x []byte) error {
	if len(x) != 1 {
		return BadLengthError
	}
	*v = Uint8View(x[0])
	return nil
}

func (v Uint8View) HashTreeRoot(HashFn) (out Root) {
	out[0] = uint8(v)
	return
}

func (v Uint8View) MarshalText() (out []byte, err error) {
	out = strconv.AppendUint(out, uint64(v), 10)
	return
}

func (v *Uint8View) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 0, 8)
	if err != nil {
		return err
	}
	*v = Uint8View(n)
	return nil
}

func (v Uint8View) MarshalJSON() ([]byte, error) {
	return conv.Uint8Marshal(uint8(v))
}

func (v *Uint8View) UnmarshalJSON(b []byte) error {
	return conv.Uint8Unmarshal((*uint8)(v), b)
}

func (v Uint8View) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

// ByteType is an alias to Uint8Type
type ByteType = Uint8Type

// ByteView is an alias to Uint8View
type ByteView = Uint8View

type Uint16View uint16

var _ BasicView = (*Uint16View)(nil)

func (v *Uint16View) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		*v = Uint16View(binary.LittleEndian.Uint16(r[:2]))
		return nil
	} else {
		return BasicBackingRequiredErr
	}
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

func (v Uint16View) Copy() (Uint16View, error) {
	return v, nil
}

func (v Uint16View) ValueByteLength() (uint64, error) {
	return 2, nil
}

func (v Uint16View) ByteLength() uint64 {
	return 2
}

func (v Uint16View) FixedLength() uint64 {
	return 2
}

func (v Uint16View) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint16(uint16(v))
}

func (v Uint16View) Encode() ([]byte, error) {
	var out [2]byte
	binary.LittleEndian.PutUint16(out[:], uint16(v))
	return out[:], nil
}

func (v *Uint16View) Deserialize(r *codec.DecodingReader) error {
	d, err := r.ReadUint16()
	if err != nil {
		return err
	}
	*v = Uint16View(d)
	return nil
}

func (v *Uint16View) Decode(x []byte) error {
	if len(x) != 2 {
		return BadLengthError
	}
	*v = Uint16View(binary.LittleEndian.Uint16(x[:2]))
	return nil
}

func (v Uint16View) HashTreeRoot(HashFn) (out Root) {
	binary.LittleEndian.PutUint16(out[:], uint16(v))
	return
}

func (v Uint16View) MarshalText() (out []byte, err error) {
	out = strconv.AppendUint(out, uint64(v), 10)
	return
}

func (v *Uint16View) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 0, 16)
	if err != nil {
		return err
	}
	*v = Uint16View(n)
	return nil
}

func (v Uint16View) MarshalJSON() ([]byte, error) {
	return conv.Uint16Marshal(uint16(v))
}

func (v *Uint16View) UnmarshalJSON(b []byte) error {
	return conv.Uint16Unmarshal((*uint16)(v), b)
}

func (v Uint16View) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

type Uint32View uint32

var _ BasicView = (*Uint32View)(nil)

func (v *Uint32View) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		*v = Uint32View(binary.LittleEndian.Uint32(r[:4]))
		return nil
	} else {
		return BasicBackingRequiredErr
	}
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

func (v Uint32View) Copy() (Uint32View, error) {
	return v, nil
}

func (v Uint32View) ValueByteLength() (uint64, error) {
	return 4, nil
}

func (v Uint32View) ByteLength() uint64 {
	return 4
}

func (v Uint32View) FixedLength() uint64 {
	return 4
}

func (v Uint32View) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint32(uint32(v))
}

func (v Uint32View) Encode() ([]byte, error) {
	var out [4]byte
	binary.LittleEndian.PutUint32(out[:], uint32(v))
	return out[:], nil
}

func (v *Uint32View) Deserialize(r *codec.DecodingReader) error {
	d, err := r.ReadUint32()
	if err != nil {
		return err
	}
	*v = Uint32View(d)
	return nil
}

func (v *Uint32View) Decode(x []byte) error {
	if len(x) != 4 {
		return BadLengthError
	}
	*v = Uint32View(binary.LittleEndian.Uint32(x[:]))
	return nil
}

func (v Uint32View) HashTreeRoot(HashFn) (out Root) {
	binary.LittleEndian.PutUint32(out[:], uint32(v))
	return out
}

func (v Uint32View) MarshalText() (out []byte, err error) {
	out = strconv.AppendUint(out, uint64(v), 10)
	return
}

func (v *Uint32View) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 0, 32)
	if err != nil {
		return err
	}
	*v = Uint32View(n)
	return nil
}

func (v Uint32View) MarshalJSON() ([]byte, error) {
	return conv.Uint32Marshal(uint32(v))
}

func (v *Uint32View) UnmarshalJSON(b []byte) error {
	return conv.Uint32Unmarshal((*uint32)(v), b)
}

func (v Uint32View) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

type Uint64View uint64

var _ BasicView = (*Uint64View)(nil)

func (v *Uint64View) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		*v = Uint64View(binary.LittleEndian.Uint64(r[:8]))
		return nil
	} else {
		return BasicBackingRequiredErr
	}
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

func (v Uint64View) Copy() (Uint64View, error) {
	return v, nil
}

func (v Uint64View) ValueByteLength() (uint64, error) {
	return 8, nil
}

func (v Uint64View) ByteLength() uint64 {
	return 8
}

func (v Uint64View) FixedLength() uint64 {
	return 8
}

func (v Uint64View) Serialize(w *codec.EncodingWriter) error {
	return w.WriteUint64(uint64(v))
}

func (v Uint64View) Encode() ([]byte, error) {
	var out [8]byte
	binary.LittleEndian.PutUint64(out[:], uint64(v))
	return out[:], nil
}

func (v *Uint64View) Deserialize(r *codec.DecodingReader) error {
	d, err := r.ReadUint64()
	if err != nil {
		return err
	}
	*v = Uint64View(d)
	return nil
}

func (v *Uint64View) Decode(x []byte) error {
	if len(x) != 8 {
		return BadLengthError
	}
	*v = Uint64View(binary.LittleEndian.Uint64(x[:]))
	return nil
}

func (v Uint64View) HashTreeRoot(HashFn) (out Root) {
	binary.LittleEndian.PutUint64(out[:], uint64(v))
	return
}

func (v Uint64View) MarshalText() (out []byte, err error) {
	out = strconv.AppendUint(out, uint64(v), 10)
	return
}

func (v *Uint64View) UnmarshalText(b []byte) error {
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = Uint64View(n)
	return nil
}

func (v Uint64View) MarshalJSON() ([]byte, error) {
	return conv.Uint64Marshal(uint64(v))
}

func (v *Uint64View) UnmarshalJSON(b []byte) error {
	return conv.Uint64Unmarshal((*uint64)(v), b)
}

func (v Uint64View) String() string {
	return strconv.FormatUint(uint64(v), 10)
}

type BoolType struct{}

var _ PackingType[BoolView] = BoolType{}

func (BoolType) New() MutView {
	return new(BoolView)
}

func (BoolType) PackViews(views []BoolView) ([]Node, error) {
	out := make([]Node, (len(views)+31)/32)
	j := 0
	for i := range out {
		var x Root
		for h := 0; h < 32 && j < len(views); j++ {
			if views[j] {
				x[h] = 1
			}
			h++
		}
		out[i] = &x
	}
	return out, nil
}

func (BoolType) SubViewFromBacking(v *Root, i uint8) BoolView {
	if i >= 32 {
		return false
	}
	if v[i] > 1 {
		return false
	}
	return v[i] == 1
}

func (BoolType) BoolViewFromBitfieldBacking(v *Root, i uint8) (BoolView, error) {
	if i > 32 {
		return false, fmt.Errorf("out of range bit lookup in node: index: %d root: %x", i, v)
	}
	return (v[i>>3]>>(i&7))&1 == 1, nil
}

func (BoolType) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (BoolType) ViewFromBacking(node Node, _ BackingHook) (BoolView, error) {
	v, ok := node.(*Root)
	if !ok {
		return false, fmt.Errorf("node %v must be a root to read a bool from it", node)
	}
	return v[0] != 0, nil
}

func (BoolType) IsFixedByteLength() bool {
	return true
}

func (BoolType) TypeByteLength() uint64 {
	return 1
}

func (BoolType) MinByteLength() uint64 {
	return 1
}

func (BoolType) MaxByteLength() uint64 {
	return 1
}

func (BoolType) Deserialize(dr *codec.DecodingReader) (BoolView, error) {
	b, err := dr.ReadByte()
	if err != nil {
		return false, err
	}
	if b > 1 {
		return false, fmt.Errorf("invalid bool value: 0x%x", b)
	}
	return b == 1, nil
}

func (BoolType) String() string {
	return "bool"
}

type BoolView bool

var _ BasicView = (*BoolView)(nil)

var trueRoot = &Root{1}

func (v *BoolView) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		*v = r[0] == 1
		return nil
	} else {
		return BasicBackingRequiredErr
	}
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
	newRoot[i] = v.byte()
	return &newRoot
}

func (v BoolView) Copy() (BoolView, error) {
	return v, nil
}

func (v BoolView) ByteLength() uint64 {
	return 1
}

func (v BoolView) ValueByteLength() (uint64, error) {
	return 1, nil
}

func (v BoolView) FixedLength() uint64 {
	return 1
}

func (v BoolView) Serialize(w *codec.EncodingWriter) error {
	return w.WriteByte(v.byte())
}

func (v BoolView) Encode() ([]byte, error) {
	return []byte{v.byte()}, nil
}

func (v *BoolView) Deserialize(r *codec.DecodingReader) error {
	d, err := r.ReadByte()
	if err != nil {
		return err
	}
	if d > 1 {
		return fmt.Errorf("invalid bool value: 0x%x", d)
	}
	*v = d > 0
	return nil
}

func (v *BoolView) Decode(x []byte) error {
	if len(x) != 1 {
		return BadLengthError
	}
	if x[0] > 1 {
		return fmt.Errorf("invalid bool value: 0x%x", x[0])
	}
	*v = x[0] > 0
	return nil
}

func (v BoolView) HashTreeRoot(HashFn) (out Root) {
	out[0] = v.byte()
	return
}

func (v BoolView) String() string {
	if v {
		return "true"
	} else {
		return "false"
	}
}
