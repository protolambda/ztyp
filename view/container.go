package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type FieldDef struct {
	Name string
	Type TypeDef
}

type ContainerType []FieldDef

func (td *ContainerType) DefaultNode() Node {
	fieldCount := td.FieldCount()
	depth := GetDepth(fieldCount)
	inner := &Commit{}
	nodes := make([]Node, fieldCount, fieldCount)
	for i, f := range *td {
		nodes[i] = f.Type.DefaultNode()
	}
	inner.ExpandInplaceTo(nodes, depth)
	return inner
}

func (td *ContainerType) ViewFromBacking(node Node, hook ViewHook) (View, error) {
	fieldCount := td.FieldCount()
	depth := GetDepth(fieldCount)
	return &ContainerView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth,
		},
		ContainerType: td,
		ViewHook: hook,
	}, nil
}

func (td *ContainerType) New(hook ViewHook) *ContainerView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*ContainerView)
}

func (td *ContainerType) FieldCount() uint64 {
	return uint64(len(*td))
}

type ContainerView struct {
	SubtreeView
	*ContainerType
	ViewHook
}

func (tv *ContainerView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *ContainerView) Get(i uint64) (View, error) {
	if count := tv.ContainerType.FieldCount(); i >= count {
		return nil, fmt.Errorf("cannot get item at field index %d, container only has %d fields", i, count)
	}
	v, err := tv.SubtreeView.Get(i)
	if err != nil {
		return nil, err
	}
	return (*tv.ContainerType)[i].Type.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ContainerView) Set(i uint64, v View) error {
	if fieldCount := tv.ContainerType.FieldCount(); i >= fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, fieldCount)
	}
	if err := tv.SubtreeView.Set(i, v.Backing()); err != nil {
		return err
	}
	return tv.PropagateChange(tv)
}

func (tv *ContainerView) ItemHook(i uint64) ViewHook {
	return func(v View) error {
		return tv.Set(i, v)
	}
}
