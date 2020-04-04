package props

import (
	"fmt"
	. "github.com/protolambda/ztyp/view"
)

// This file is why we need generics in Go. Could have been ~10 lines.

type ContainerProp ReadPropFn

func (p ContainerProp) Container() (*ContainerView, error) {
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

type ComplexVectorProp ReadPropFn

func (p ComplexVectorProp) Vector() (*ComplexVectorView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ComplexVectorView)
	if ok {
		return nil, fmt.Errorf("view is not a vector: %v", v)
	}
	return c, nil
}

type BasicVectorProp ReadPropFn

func (p BasicVectorProp) BasicVector() (*BasicVectorView, error) {
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

type BitVectorProp ReadPropFn

func (p BitVectorProp) BitVector() (*BitVectorView, error) {
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

type ComplexListProp ReadPropFn

func (p ComplexListProp) List() (*ComplexListView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	c, ok := v.(*ComplexListView)
	if ok {
		return nil, fmt.Errorf("view is not a list: %v", v)
	}
	return c, nil
}

type BasicListProp ReadPropFn

func (p BasicListProp) BasicList() (*BasicListView, error) {
	v, err := p()
	if err != nil {
		return nil, err
	}
	bv, ok := v.(*BasicListView)
	if ok {
		return nil, fmt.Errorf("view is not a basic list: %v", v)
	}
	return bv, nil
}

type BitListProp ReadPropFn

func (p BitListProp) BitList() (*BitListView, error) {
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
