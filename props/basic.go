package props

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
)

type RootReadProp ReadPropFn

func (p RootReadProp) Root() (Root, error) {
	v, err := p()
	if err != nil {
		return Root{}, err
	}
	r, ok := v.(*RootView)
	if ok {
		return Root{}, fmt.Errorf("not a root view: %v", v)
	}
	return Root(*r), nil
}

type RootWriteProp WritePropFn

func (p RootWriteProp) SetRoot(v Root) error {
	rv := RootView(v)
	return p(&rv)
}

type Uint64ReadProp ReadPropFn

func (p Uint64ReadProp) Uint64() (uint64, error) {
	v, err := p()
	if err != nil {
		return 0, err
	}
	n, ok := v.(Uint64View)
	if ok {
		return 0, fmt.Errorf("not a uint64 view: %v", v)
	}
	return uint64(n), nil
}

type Uint64WriteProp WritePropFn

func (p Uint64WriteProp) SetUint64(v uint64) error {
	return p(Uint64View(v))
}

type BoolReadProp ReadPropFn

func (p BoolReadProp) Bool() (bool, error) {
	v, err := p()
	if err != nil {
		return false, err
	}
	b, ok := v.(BoolView)
	if ok {
		return false, fmt.Errorf("not a bool view: %v", v)
	}
	return bool(b), nil
}

type BoolWriteProp WritePropFn

func (p BoolWriteProp) SetBool(v BoolView) error {
	return p(v)
}
