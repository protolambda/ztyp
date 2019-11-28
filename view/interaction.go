package view

type ReadProp func() (View, error)

type WriteProp func(v View) error

type MutProp struct {
	ReadProp
	WriteProp
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

func PropReader(rv ReadablePropView, i uint64) ReadProp {
	return func() (View, error) {
		return rv.Get(i)
	}
}

func PropWriter(wv WritablePropView, i uint64) WriteProp {
	return func(v View) error {
		return wv.Set(i, v)
	}
}

func PropMutator(mv MutablePropView, i uint64) MutProp {
	return MutProp{
		ReadProp:  PropReader(mv, i),
		WriteProp: PropWriter(mv, i),
	}
}
