package props

import (
	"errors"
	. "github.com/protolambda/ztyp/view"
)

type ReadProp interface {
	Read() (View, error)
}

type ReadPropFn func() (View, error)

func (f ReadPropFn) Read() (View, error) {
	if f == nil {
		return nil, errors.New("property is not available to read")
	}
	return f()
}

type WriteProp interface {
	Write(v View) error
}

type WritePropFn func(v View) error

func (f WritePropFn) Write(v View) error {
	if f == nil {
		return errors.New("property is not available to write")
	}
	return f(v)
}

type MutProp interface {
	ReadProp
	WriteProp
}

type MutPropFns struct {
	ReadPropFn
	WritePropFn
}

type ReadablePropView interface {
	Get(i uint64) (View, error)
}

type WritablePropView interface {
	Set(i uint64, v View) error
}

func PropReader(rv ReadablePropView, i uint64) ReadPropFn {
	return func() (View, error) {
		return rv.Get(i)
	}
}

func PropWriter(wv WritablePropView, i uint64) WritePropFn {
	return func(v View) error {
		return wv.Set(i, v)
	}
}
