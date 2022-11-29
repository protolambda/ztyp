package view

import (
	"fmt"

	"github.com/protolambda/ztyp/codec"
	. "github.com/protolambda/ztyp/tree"
)

const OffsetByteLength = 4

type ComplexTypeBase struct {
	MinSize     uint64
	MaxSize     uint64
	Size        uint64
	IsFixedSize bool
}

func (td *ComplexTypeBase) IsFixedByteLength() bool {
	return td.IsFixedSize
}

func (td *ComplexTypeBase) TypeByteLength() uint64 {
	return td.Size
}

func (td *ComplexTypeBase) MinByteLength() uint64 {
	return td.MinSize
}

func (td *ComplexTypeBase) MaxByteLength() uint64 {
	return td.MaxSize
}

func (td *ComplexTypeBase) checkScope(scope uint64) error {
	if scope < td.MinSize {
		return fmt.Errorf("scope %d is too small, need at least %d bytes", scope, td.MinSize)
	}
	if scope > td.MaxSize {
		return fmt.Errorf("scope %d is too big, need %d or less bytes", scope, td.MaxSize)
	}
	return nil
}

type ErrNodeIter struct {
	error
}

func (e ErrNodeIter) Next() (chunk Node, ok bool, err error) {
	return nil, false, e.error
}

type NodeIterFn func() (chunk Node, ok bool, err error)

func (f NodeIterFn) Next() (chunk Node, ok bool, err error) {
	return f()
}

type NodeIter interface {
	// Next gets the next node, ok is true if it actually exists.
	// An error may occur if data is missing or corrupt.
	Next() (chunk Node, ok bool, err error)
}

type ErrElemIter[EV View, ET TypeDef[EV]] struct {
	error
}

func (e ErrElemIter[EV, ET]) Next() (elem EV, elemTyp ET, ok bool, err error) {
	err = e.error
	return
}

type ElemIterFn[EV View, ET TypeDef] func(elem EV) (elemTyp ET, ok bool, err error)

func (f ElemIterFn[EV, ET]) Next(elem EV) (elemTyp ET, ok bool, err error) {
	return (func(elem EV) (elemTyp ET, ok bool, err error))(f)(elem)
}

type ElemIter[EV MutView, ET TypeDef] interface {
	// Next gets the next element, ok is true if it actually exists.
	// An error may occur if data is missing or corrupt.
	Next(elem EV) (elemTyp ET, ok bool, err error)
}

func serializeComplexFixElemSeries[EV View, ET TypeDef[EV]](iter ElemIter[EV, ET], w *codec.EncodingWriter) error {
	for {
		el, _, ok, err := iter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if err := el.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}

func serializeComplexVarElemSeries[EV View, ET TypeDef[EV]](length uint64, iterFn func() ElemIter[EV, ET], w *codec.EncodingWriter) error {
	// the previous offset, to calculate a new offset from, starting after the fixed data.
	prevOffset := length * OffsetByteLength

	// span of the previous var-size element
	prevSize := uint64(0)
	iter := iterFn()
	// write all offsets
	for {
		el, _, ok, err := iter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		elValSize, err := el.ValueByteLength()
		if err != nil {
			return err
		}
		prevOffset, err = w.WriteOffset(prevOffset, prevSize)
		if err != nil {
			return err
		}
		prevSize = elValSize
	}
	iter = iterFn()
	// now write all elements
	for {
		el, _, ok, err := iter.Next()
		if err != nil {
			return err
		}
		if !ok {
			break
		}
		if err := el.Serialize(w); err != nil {
			return err
		}
	}
	return nil
}
