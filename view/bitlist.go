package view

import (
	"encoding/binary"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
)

type BitListTypeDef struct {
	BitLimit uint64
}

func (td *BitListTypeDef) DefaultNode() Node {
	depth := GetDepth(td.BottomNodeLimit())
	return &Commit{Left: &ZeroHashes[depth], Right: &ZeroHashes[0]}
}

func (td *BitListTypeDef) ViewFromBacking(node Node, hook ViewHook) (View, error) {
	depth := GetDepth(td.BottomNodeLimit())
	return &BitListView{
		SubtreeView: SubtreeView{
			BackingNode: node,
			depth:       depth + 1, // +1 for length mix-in
		},
		BitListTypeDef: td,
		ViewHook:       hook,
	}, nil
}

func (td *BitListTypeDef) BottomNodeLimit() uint64 {
	return (td.BitLimit + 0xff) >> 8
}

func (td *BitListTypeDef) New(hook ViewHook) *BitListView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*BitListView)
}

func BitlistType(limit uint64) *BitListTypeDef {
	return &BitListTypeDef{
		BitLimit: limit,
	}
}

type BitListView struct {
	SubtreeView
	*BitListTypeDef
	ViewHook
}

func (tv *BitListView) ViewRoot(h HashFn) Root {
	return tv.BackingNode.MerkleRoot(h)
}

func (tv *BitListView) Append(view BoolView) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll >= tv.BitLimit {
		return fmt.Errorf("list length is %d and appending would exceed the list limit %d", ll, tv.BitLimit)
	}
	// Appending is done by modifying the bottom node at the index list_length. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64(ll>>8, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(lastGindex)
	if err != nil {
		return fmt.Errorf("failed to get a setter to append an item")
	}
	if ll&0xff == 0 {
		// New bottom node
		tv.BackingNode = setLast(view.BackingFromBitfieldBase(&ZeroHashes[0], 0))
	} else {
		// Apply to existing partially zeroed bottom node
		r, _, subIndex, err := tv.subviewNode(ll)
		if err != nil {
			return err
		}
		tv.BackingNode = setLast(view.BackingFromBitfieldBase(r, subIndex))
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

func (tv *BitListView) Pop() error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if ll == 0 {
		return fmt.Errorf("list length is 0 and no bit can be popped")
	}
	// Popping is done by modifying the bottom node at the index list_length - 1. And expanding where necessary as it is being set.
	lastGindex, err := ToGindex64((ll-1)>>8, tv.depth)
	if err != nil {
		return err
	}
	setLast, err := tv.SubtreeView.BackingNode.ExpandInto(lastGindex)
	if err != nil {
		return fmt.Errorf("failed to get a setter to pop a bit")
	}
	// Get the subview to erase
	r, _, subIndex, err := tv.subviewNode(ll - 1)
	if err != nil {
		return err
	}
	// Pop the bit by setting it to the default
	// Update the view to the new tree containing this item.
	tv.BackingNode = setLast(BoolView(false).BackingFromBitfieldBase(r, subIndex))
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

func (tv *BitListView) CheckIndex(i uint64) error {
	ll, err := tv.Length()
	if err != nil {
		return err
	}
	if i >= ll {
		return fmt.Errorf("cannot handle item at element index %d, list only has %d bits", i, ll)
	}
	if i >= tv.BitLimit {
		return fmt.Errorf("bitlist has a an invalid length of %d and cannot handle a bit at index %d because of a limit of %d bits", ll, i, tv.BitLimit)
	}
	return nil
}

func (tv *BitListView) subviewNode(i uint64) (r *Root, bottomIndex uint64, subIndex uint8, err error) {
	bottomIndex, subIndex = i>>8, uint8(i)
	v, err := tv.SubtreeView.Get(bottomIndex)
	if err != nil {
		return nil, 0, 0, err
	}
	r, ok := v.(*Root)
	if !ok {
		return nil, 0, 0, fmt.Errorf("bitlist bottom node is not a root, at index %d", i)
	}
	return r, bottomIndex, subIndex, nil
}

func (tv *BitListView) Get(i uint64) (BoolView, error) {
	if err := tv.CheckIndex(i); err != nil {
		return false, err
	}
	r, _, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return false, err
	}
	return BoolType.BoolViewFromBitfieldBacking(r, subIndex)
}

func (tv *BitListView) Set(i uint64, v BoolView) error {
	if err := tv.CheckIndex(i); err != nil {
		return err
	}
	r, bottomIndex, subIndex, err := tv.subviewNode(i)
	if err != nil {
		return err
	}
	if err := tv.SubtreeView.Set(bottomIndex, v.BackingFromBitfieldBase(r, subIndex)); err != nil {
		return err
	}
	return tv.PropagateChange(tv)
}

func (tv *BitListView) Length() (uint64, error) {
	v, err := tv.SubtreeView.BackingNode.Getter(RightGindex)
	if err != nil {
		return 0, err
	}
	llBytes, ok := v.(*Root)
	if !ok {
		return 0, fmt.Errorf("cannot read node %v as list-length", v)
	}
	ll := binary.LittleEndian.Uint64(llBytes[:8])
	if ll > tv.BitLimit {
		return 0, fmt.Errorf("cannot read list length, length appears to be bigger than limit allows")
	}
	return ll, nil
}
