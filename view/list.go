package view

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type ListTypeDef struct {
	ElementType TypeDef
	Limit       uint64
}

func (td *ListTypeDef) DefaultNode() Node {
	depth := GetDepth(td.Limit)
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (td *ListTypeDef) ViewFromBacking(node Node) (View, error) {
	depth := GetDepth(td.Limit)
	return &ListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		ListTypeDef: td,
	}, nil
}

func (td *ListTypeDef) New() *ListView {
	v, _ := td.ViewFromBacking(td.DefaultNode())
	return v.(*ListView)
}

func ListType(elemType TypeDef, limit uint64) *ListTypeDef {
	return &ListTypeDef{
		ElementType: elemType,
		Limit:       limit,
	}
}

type ListView struct {
	SubtreeView
	*ListTypeDef
}

func (tv *ListView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *ListView) Append(v Node) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll >= tv.Limit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, tv.Limit)
	}
	// Appending is done by setting the node at the index list_length. And expanding where necessary as it is being set.
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(ll, tv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	// Append the item by setting the newly allocated last item to it.
	// Update the view to the new tree containing this item.
	tv.BackingNode = setLast(v)
	// And update the list length
	setLength, err := tv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll+1)
	tv.BackingNode = setLength(newLength)
	return nil
}

func (tv *ListView) Pop() error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no item can be popped")
	}
	// Popping is done by setting the node at the index list_length - 1. And expanding where necessary as it is being set.
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(ll-1, tv.depth)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item")
	}
	// Pop the item by setting it to the zero hash
	// Update the view to the new tree containing this item.
	tv.BackingNode = setLast(&ZeroHashes[0])
	// And update the list length
	setLength, err := tv.SubtreeView.BackingNode.Setter(1, 1)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll-1)
	tv.BackingNode = setLength(newLength)
	return nil
}

func (tv *ListView) CheckIndex(i uint64) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot handle item at element index %d, list only has %d elements", i, ll)
	}
	if i >= tv.Limit {
		return fmt.Errorf("list has a an invalid length of %d and cannot handle an element at index %d because of a limit of %d elements", ll, i, tv.Limit)
	}
	return nil
}

func (tv *ListView) Get(i uint64) (Node, error) {
	if err := tv.CheckIndex(i); err != nil {
		return nil, err
	}
	return tv.SubtreeView.Get(i)
}

func (tv *ListView) Set(i uint64, node Node) error {
	if err := tv.CheckIndex(i); err != nil {
		return err
	}
	return tv.SubtreeView.Set(i, node)
}

func (tv *ListView) Length() (uint64, error) {
	v, err := tv.SubtreeView.BackingNode.Getter(1, 1)
	if err != nil {
		return 0, err
	}
	llBytes, ok := v.(*Root)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	ll := binary.LittleEndian.Uint64(llBytes[:8])
	if ll > tv.Limit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return ll, nil
}
