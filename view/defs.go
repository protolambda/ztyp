package view

import . "github.com/protolambda/ztyp/tree"

type View interface {
	Backing() Node
}

type ViewHook func(v View) error

func (vh ViewHook) PropagateChange(v View) error {
	if vh != nil {
		return vh(v)
	} else {
		return nil
	}
}

type TypeDef interface {
	DefaultNode() Node
	ViewFromBacking(node Node, hook ViewHook) (View, error)
}

type SubView interface {
	BackingFromBase(base *Root, i uint8) *Root
}

type BasicTypeDef interface {
	TypeDef
	ByteLength() uint64
	SubViewFromBacking(node *Root, i uint8) (SubView, error)
}
