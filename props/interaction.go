package props

import . "github.com/protolambda/ztyp/view"

type ReadProp interface {
	Read() (View, error)
}

type ReadPropFn func() (View, error)

func (f ReadPropFn) Read() (View, error) {
	return f()
}

type WriteProp interface {
	Write(v View) error
}

type WritePropFn func(v View) error

func (f WritePropFn) Write(v View) error {
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

type MutablePropView interface {
	ReadablePropView
	WritablePropView
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

func PropMutator(mv MutablePropView, i uint64) MutPropFns {
	return MutPropFns{
		ReadPropFn:  PropReader(mv, i),
		WritePropFn: PropWriter(mv, i),
	}
}
