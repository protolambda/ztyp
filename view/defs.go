package view

import (
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type View interface {
	Backing() Node
	SetBacking(b Node) error
	Copy() (View, error)
	ValueByteLength() uint64
	Serialize(w io.Writer) error
	HashTreeRoot(h HashFn) Root
}

type ViewBase struct {
	TypeDef TypeDef
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
	Default(hook BackingHook) View
	DefaultNode() Node
	ViewFromBacking(node Node, hook BackingHook) (View, error)
	IsFixedByteLength() bool
	// 0 if there type has no single fixed byte length
	TypeByteLength() uint64
	MinByteLength() uint64
	MaxByteLength() uint64
	Deserialize(r io.Reader, scope uint64) error
	Name() string
	String() string
	// TODO: could add navigation by key/index into subtypes
}

type BasicView interface {
	BackingFromBase(base *Root, i uint8) *Root
}

type BasicTypeDef interface {
	TypeDef
	BasicViewFromBacking(node *Root, i uint8) (BasicView, error)
	PackViews(views []BasicView) ([]Node, error)
}
