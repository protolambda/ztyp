package view

import (
	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

type View interface {
	Backing() Node
	SetBacking(b Node) error
	ValueByteLength() (uint64, error)
	Serialize(w *codec.EncodingWriter) error
	HashTreeRoot(h HashFn) Root
}

type TypedView interface {
	View
	Type() TypeDef[View]
}

type BackingHook func(b Node) error

func (vh BackingHook) PropagateChangeMaybe(b Node) error {
	if vh != nil {
		return vh(b)
	} else {
		return nil
	}
}

type TypeDef[V View] interface {
	Default(hook BackingHook) V
	Mask() TypeDef[View]
	DefaultNode() Node
	ViewFromBacking(node Node, hook BackingHook) (V, error)
	IsFixedByteLength() bool
	// 0 if there type has no single fixed byte length
	TypeByteLength() uint64
	MinByteLength() uint64
	MaxByteLength() uint64
	Deserialize(dr *codec.DecodingReader) (V, error)
	String() string
	// TODO: could add navigation by key/index into subtypes
}

type BasicView interface {
	View
	BackingFromBase(base *Root, i uint8) *Root
}

type BasicTypeDef[V BasicView] interface {
	TypeDef[V]
	BasicViewFromBacking(node *Root, i uint8) (V, error)
	PackViews(views []V) ([]Node, error)
}
