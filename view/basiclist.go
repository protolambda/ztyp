package view

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type BasicListTypeDef struct {
	ElementType BasicTypeDef
	Limit       uint64
}

func (td *BasicListTypeDef) DefaultNode() Node {
	depth := GetDepth(td.BottomNodeLimit())
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (td *BasicListTypeDef) ViewFromBacking(node Node, hook ViewHook) (View, error) {
	depth := GetDepth(td.BottomNodeLimit())
	return &BasicListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		BasicListTypeDef: td,
		ViewHook:         hook,
	}, nil
}

func (td *BasicListTypeDef) ElementsPerBottomNode() uint64 {
	return 32 / td.ElementType.ByteLength()
}

func (td *BasicListTypeDef) BottomNodeLimit() uint64 {
	perNode := td.ElementsPerBottomNode()
	return (td.Limit + perNode - 1) / perNode
}

func (td *BasicListTypeDef) TranslateIndex(index uint64) (nodeIndex uint64, intraNodeIndex uint8) {
	perNode := td.ElementsPerBottomNode()
	return index / perNode, uint8(index & (perNode - 1))
}

func (td *BasicListTypeDef) New(hook ViewHook) *BasicListView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*BasicListView)
}

func BasicListType(elemType BasicTypeDef, limit uint64) *BasicListTypeDef {
	return &BasicListTypeDef{
		ElementType: elemType,
		Limit:       limit,
	}
}

type BasicListView struct {
	SubtreeView
	*BasicListTypeDef
	ViewHook
}

func (tv *BasicListView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *BasicListView) Append(view SubView) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll >= tv.Limit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, tv.Limit)
	}
	perNode := tv.ElementsPerBottomNode()
	// Appending is done by modifying the bottom node at the index list_length. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64(ll/perNode, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(lastGindex)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	if ll%perNode == 0 {
		// New bottom node
		tv.BackingNode = setLast(view.BackingFromBase(&ZeroHashes[0], 0))
	} else {
		// Apply to existing partially zeroed bottom node
		r, _, subIndex, err := tv.subviewNode(ll)
		if err != nil {
			return err
		}
		tv.BackingNode = setLast(view.BackingFromBase(r, subIndex))
	}
	// And update the list length
	setLength, err := tv.SubtreeView.BackingNode.Setter(RightGindex)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll+1)
	tv.BackingNode = setLength(newLength)
	return tv.PropagateChange(tv)
}

func (tv *BasicListView) Pop() error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no item can be popped")
	}
	perNode := tv.ElementsPerBottomNode()
	// Popping is done by modifying the bottom node at the index list_length - 1. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64((ll-1)/perNode, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(lastGindex)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop an item")
	}
	// Get the subview to erase
	r, _, subIndex, err := tv.subviewNode(ll - 1)
	if err != nil {
		return err
	}
	// Pop the item by setting it to the default
	// Update the view to the new tree containing this item.
	defaultElement, err := tv.ElementType.SubViewFromBacking(&ZeroHashes[0], subIndex)
	if err != nil {
		return err
	}
	tv.BackingNode = setLast(defaultElement.BackingFromBase(r, subIndex))
	// And update the list length
	setLength, err := tv.SubtreeView.BackingNode.Setter(RightGindex)
	if err != nil {
		return err
	}
	newLength := &Root{}
	binary.LittleEndian.PutUint64(newLength[:8], ll-1)
	tv.BackingNode = setLength(newLength)
	return tv.PropagateChange(tv)
}

func (tv *BasicListView) CheckIndex(i uint64) error {
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

func (tv *BasicListView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = tv.TranslateIndex(i)
	v, err := tv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("basic list bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (tv *BasicListView) Get(i uint64) (SubView, error) {
	if err := tv.CheckIndex(i); err != nil {
		return nil, err
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ElementType.SubViewFromBacking(r, subIndex)
}

func (tv *BasicListView) Set(i uint64, v SubView) error {
	if err := tv.CheckIndex(i); err != nil {
		return err
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	if err := tv.SubtreeView.Set(bottomIndex, v.BackingFromBase(r, subIndex)); err != nil {
		return err
	}
	return tv.PropagateChange(tv)
}

func (tv *BasicListView) Length() (uint64, error) {
	v, err := tv.SubtreeView.BackingNode.Getter(RightGindex)
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
