package view

import (
	"encoding/binary"
	"fmt"
	"math/big"
	"strconv"

	"github.com/holiman/uint256"

	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/conv"
	. "github.com/protolambda/ztyp/tree"
)

type Uint256Type struct{}

var _ PackingType[Uint256View] = Uint256Type{}

func (Uint256Type) DefaultNode() Node {
	return &ZeroHashes[0]
}

func (Uint256Type) New() MutView {
	return new(Uint256View)
}

func (Uint256Type) PackViews(views []Uint256View) ([]Node, error) {
	out := make([]Node, len(views))
	for i := range out {
		x := Root(views[i].Bytes32())
		out[i] = &x
	}
	return out, nil
}

func (Uint256Type) IsFixedByteLength() bool {
	return true
}

func (Uint256Type) TypeByteLength() uint64 {
	return 32
}

func (Uint256Type) MinByteLength() uint64 {
	return 32
}

func (Uint256Type) MaxByteLength() uint64 {
	return 32
}

func (Uint256Type) Deserialize(dr *codec.DecodingReader) (Uint256View, error) {
	var out Uint256View
	err := out.Deserialize(dr)
	return out, err
}

func (td Uint256Type) String() string {
	return "uint256"
}

type Uint256View uint256.Int

var _ BasicView = (*Uint256View)(nil)

func (v *Uint256View) SetBacking(b Node) error {
	if r, ok := b.(*Root); ok {
		v.setBytes32(r[:])
		return nil
	} else {
		return BasicBackingRequiredErr
	}
}

// Bytes32 returns little endian encoding
func (v Uint256View) Bytes32() (out [32]byte) {
	// uint64 3 is most significant internally in uint256.Int
	binary.LittleEndian.PutUint64(out[0:8], v[0])
	binary.LittleEndian.PutUint64(out[8:16], v[1])
	binary.LittleEndian.PutUint64(out[16:24], v[2])
	binary.LittleEndian.PutUint64(out[24:32], v[3])
	return
}

// SetBytes32 sets view from little endian encoding
func (v *Uint256View) SetBytes32(data [32]byte) {
	v.setBytes32(data[:])
}

func (v *Uint256View) setBytes32(data []byte) {
	v[0] = binary.LittleEndian.Uint64(data[0:8])
	v[1] = binary.LittleEndian.Uint64(data[8:16])
	v[2] = binary.LittleEndian.Uint64(data[16:24])
	v[3] = binary.LittleEndian.Uint64(data[24:32])
}

// Bytes returns little endian encoding (always 32 bytes)
func (v Uint256View) Bytes() []byte {
	out := v.Bytes32()
	return out[:]
}

func (v Uint256View) Backing() Node {
	out := Root(v.Bytes32())
	return &out
}

func (v Uint256View) BackingFromBase(_ *Root, i uint8) *Root {
	if i != 0 { // must always be aligned, we overwrite full 32 bytes here, nothing to pack along in same root.
		return nil
	}
	newRoot := Root(v.Bytes32())
	return &newRoot
}

func (v Uint256View) Copy() (View, error) {
	return &v, nil
}

func (v Uint256View) ValueByteLength() (uint64, error) {
	return 32, nil
}

func (v Uint256View) ByteLength() uint64 {
	return 32
}

func (v Uint256View) FixedLength() uint64 {
	return 32
}

func (v Uint256View) Serialize(w *codec.EncodingWriter) error {
	return w.Write(v.Bytes())
}

func (v Uint256View) Encode() ([]byte, error) {
	return v.Bytes(), nil
}

func (v *Uint256View) Deserialize(r *codec.DecodingReader) error {
	var data [32]byte
	_, err := r.Read(data[:])
	if err != nil {
		return err
	}
	v.SetBytes32(data)
	return nil
}

func (v *Uint256View) Decode(x []byte) error {
	if len(x) != 32 {
		return BadLengthError
	}
	v.setBytes32(x)
	return nil
}

func (v Uint256View) HashTreeRoot(HashFn) Root {
	return v.Bytes32()
}

func (v Uint256View) MarshalText() (out []byte, err error) {
	return []byte(fmt.Sprintf("%d", (*uint256.Int)(&v))), nil
}

func (v *Uint256View) SetFromBig(x *big.Int) (overflow bool) {
	if x.Sign() < 0 {
		return true
	}
	return (*uint256.Int)(v).SetFromBig(x)
}

func (v *Uint256View) UnmarshalText(b []byte) error {
	x := new(big.Int)
	err := x.UnmarshalText(b)
	if err != nil {
		return fmt.Errorf("failed to unmarshal Uint256View: %w", err)
	}
	if v.SetFromBig(x) {
		return strconv.ErrRange
	}
	return nil
}

func (v Uint256View) MarshalJSON() ([]byte, error) {
	return conv.Uint256Marshal((*uint256.Int)(&v))
}

func (v *Uint256View) UnmarshalJSON(b []byte) error {
	return conv.Uint256Unmarshal((*uint256.Int)(v), b)
}

func (v Uint256View) String() string {
	return fmt.Sprintf("%d", (*uint256.Int)(&v))
}

func MustUint256(v string) Uint256View {
	var out Uint256View
	err := out.UnmarshalText([]byte(v))
	if err != nil {
		panic(err)
	}
	return out
}
