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
	r, ok := v.(*Root)
	if ok {
		return Root{}, fmt.Errorf("not a root view: %v", v)
	}
	return *r, nil
}

type RootWriteProp WritePropFn

func (p RootWriteProp) SetRoot(v *Root) error {
	return p(v)
}

type Uint64ReadProp ReadPropFn

func (p Uint64ReadProp) Uint64() (Uint64View, error) {
	v, err := p()
	if err != nil {
		return 0, err
	}
	n, ok := v.(Uint64View)
	if ok {
		return 0, fmt.Errorf("not a uint64 view: %v", v)
	}
	return n, nil
}

type Uint64WriteProp WritePropFn

func (p Uint64WriteProp) SetUint64(v Uint64View) error {
	return p(v)
}

type BoolReadProp ReadPropFn

func (p BoolReadProp) Bool() (BoolView, error) {
	v, err := p()
	if err != nil {
		return false, err
	}
	b, ok := v.(BoolView)
	if ok {
		return false, fmt.Errorf("not a bool view: %v", v)
	}
	return b, nil
}

type BoolWriteProp WritePropFn

func (p BoolWriteProp) SetBool(v BoolView) error {
	return p(v)
}
