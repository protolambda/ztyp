package props

import (
	"fmt"
	. "github.com/protolambda/ztyp/view"
)

type ContainerReadProp ReadPropFn

func (p ContainerReadProp) Container() (*ContainerView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ContainerView)
	if ok {
		return nil, fmt.Errorf("view is not a container: %v", v)
	}
	return c, nil
}

type ContainerWriteProp WritePropFn

func (p ContainerWriteProp) SetContainer(v *ContainerView) error {
	return p(v)
}

type BasicVectorReadProp ReadPropFn

func (p BasicVectorReadProp) BasicVector() (*BasicVectorView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BasicVectorView)
	if ok {
		return nil, fmt.Errorf("not a uint64 view: %v", v)
	}
	return bv, nil
}

