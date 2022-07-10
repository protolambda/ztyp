package view

import (
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
)

// Mask is an util to make any TypeDef[V] fit in a dull TypeDef[View]
type Mask[V View, T TypeDef[V]] struct {
	T T
}

var _ TypeDef[View] = Mask[Uint8View, Uint8Type]{}
var _ TypeDef[View] = Mask[*ContainerView, *ContainerTypeDef]{}

func (t Mask[V, T]) Default(hook BackingHook) View {
	return t.T.Default(hook)
}

func (t Mask[V, T]) Mask() TypeDef[View] {
	return t
}

func (t Mask[V, T]) DefaultNode() tree.Node {
	return t.T.DefaultNode()
}

func (t Mask[V, T]) ViewFromBacking(node tree.Node, hook BackingHook) (View, error) {
	return t.T.ViewFromBacking(node, hook)
}

func (t Mask[V, T]) IsFixedByteLength() bool {
	return t.T.IsFixedByteLength()
}

func (t Mask[V, T]) TypeByteLength() uint64 {
	return t.T.TypeByteLength()
}

func (t Mask[V, T]) MinByteLength() uint64 {
	return t.T.MinByteLength()
}

func (t Mask[V, T]) MaxByteLength() uint64 {
	return t.T.MaxByteLength()
}

func (t Mask[V, T]) Deserialize(dr *codec.DecodingReader) (View, error) {
	return t.T.Deserialize(dr)
}

func (t Mask[V, T]) String() string {
	return t.T.String()
}
