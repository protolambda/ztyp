package props

import (
	"fmt"
	. "github.com/protolambda/ztyp/view"
)

// This file is why we need generics in Go. Could have been ~10 lines.

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

type VectorReadProp ReadPropFn

func (p VectorReadProp) Vector() (*VectorView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	c, ok := v.(*VectorView)
	if ok {
		return nil, fmt.Errorf("view is not a vector: %v", v)
	}
	return c, nil
}

type BasicVectorReadProp ReadPropFn

func (p BasicVectorReadProp) BasicVector() (*BasicVectorView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BasicVectorView)
	if ok {
		return nil, fmt.Errorf("view is not a basic vector: %v", v)
	}
	return bv, nil
}

type BitVectorReadProp ReadPropFn

func (p BitVectorReadProp) BitVector() (*BitVectorView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BitVectorView)
	if ok {
		return nil, fmt.Errorf("view is not a bitvector: %v", v)
	}
	return bv, nil
}

type ListReadProp ReadPropFn

func (p ListReadProp) List() (*ListView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ListView)
	if ok {
		return nil, fmt.Errorf("view is not a list: %v", v)
	}
	return c, nil
}

type BitListReadProp ReadPropFn

func (p BitListReadProp) BitList() (*BitListView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BitListView)
	if ok {
		return nil, fmt.Errorf("view is not a bitlist: %v", v)
	}
	return bv, nil
}
