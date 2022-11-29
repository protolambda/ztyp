package view

import (
	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

// Backing: immutable tree representing merkle-tree presentation of an SSZ value
//
// View: mutable interface around a backing, replaces the backing on mutation.
// The new backing may use data-sharing to avoid deep copying
//
// TypeDef: defines the SSZ type structure, to create views and backings with.

type View interface {
	Backing() Node
	ValueByteLength() (uint64, error)
	Serialize(w *codec.EncodingWriter) error
	HashTreeRoot(h HashFn) Root
}

type MutView interface {
	View
	SetBacking(b Node) error
	Deserialize(dr *codec.DecodingReader) error
}

type HookedView interface {
	MutView
	SetHook(hook BackingHook)
}

type BackingHook func(b Node) error

func (vh BackingHook) PropagateChangeMaybe(b Node) error {
	if vh != nil {
		return vh(b)
	} else {
		return nil
	}
}

type TypeDef interface {
	DefaultNode() Node
	New() MutView
	IsFixedByteLength() bool
	// 0 if there type has no single fixed byte length
	TypeByteLength() uint64
	MinByteLength() uint64
	MaxByteLength() uint64
	String() string
	// TODO: could add navigation by key/index into subtypes
}

type BasicView interface {
	View
	BackingFromBase(base *Root, i uint8) *Root
}

type MutBasicView interface {
	BasicView
	MutView
	Decode(x []byte) error
}
