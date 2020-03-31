package view

import (
	"bytes"
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"io"
)

type FieldDef struct {
	Name string
	Type TypeDef
}

type ContainerType struct {
	TypeName string
	MinSize uint64
	MaxSize uint64
	Size uint64
	IsFixedSize bool
	Fields []FieldDef
}

func Container(name string, fields []FieldDef) *ContainerType {
	minSize := uint64(0)
	maxSize := uint64(0)
	size := uint64(0)
	isFixedSize := true
	for _, f := range fields {
		if f.Type.IsFixedByteLength() {
			size += f.Type.TypeByteLength()
		} else {
			isFixedSize = false
			minSize += f.Type.MinByteLength()
			maxSize += f.Type.MaxByteLength()
		}
	}
	if isFixedSize {
		minSize = size
		maxSize = size
	} else {
		size = 0
	}
	return &ContainerType{
		TypeName: name,
		MinSize: minSize,
		MaxSize: maxSize,
		Size: size,
		IsFixedSize: isFixedSize,
		Fields: fields,
	}
}

func (td *ContainerType) DefaultNode() Node {
	fieldCount := td.FieldCount()
	depth := CoverDepth(fieldCount)
	nodes := make([]Node, fieldCount, fieldCount)
	for i, f := range td.Fields {
		nodes[i] = f.Type.DefaultNode()
	}
	// can ignore error, depth is derive from nodes count.
	rootNode, _ := SubtreeFillToContents(nodes, depth)
	return rootNode
}

func (td *ContainerType) ViewFromBacking(node Node, hook BackingHook) (View, error) {
	fieldCount := td.FieldCount()
	depth := CoverDepth(fieldCount)
	return &ContainerView{
		SubtreeView: SubtreeView{
			BackedView: BackedView{
				ViewBase: ViewBase{
					TypeDef: td,
				},
				Hook: hook,
				BackingNode: node,
			},
			depth:       depth,
		},
	}, nil
}

func (td *ContainerType) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *ContainerType) New(hook BackingHook) *ContainerView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*ContainerView)
}

func (td *ContainerType) FieldCount() uint64 {
	return uint64(len(td.Fields))
}

func (td *ContainerType) IsFixedByteLength() bool {
	return td.IsFixedSize
}

func (td *ContainerType) TypeByteLength() uint64 {
	return td.Size
}

func (td *ContainerType) MinByteLength() uint64 {
	return td.MinSize
}

func (td *ContainerType) MaxByteLength() uint64 {
	return td.MaxSize
}

func (td *ContainerType) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *ContainerType) Name() string {
	return td.TypeName
}

func (td *ContainerType) TypeRepr() string {
	var buf bytes.Buffer
	buf.WriteString(td.TypeName)
	buf.WriteString("(Container):")
	for _, f := range td.Fields {
		buf.WriteString("    ")
		buf.WriteString(f.Name)
		buf.WriteString(": ")
		buf.WriteString(f.Type.Name())
		buf.WriteRune('\n')
	}
	return buf.String()
}


type ContainerView struct {
	SubtreeView
	*ContainerType
}

func (tv *ContainerView) Copy() (View, error) {
	tvCopy := *tv
	tvCopy.Hook = nil
	return &tvCopy, nil
}

func (tv *ContainerView) ValueByteLength() uint64 {
	if tv.IsFixedSize {
		return tv.Size
	}
	// TODO
	return 0
}

func (tv *ContainerView) Serialize(w io.Writer) error {
	// TODO
	return nil
}

func (tv *ContainerView) Get(i uint64) (View, error) {
	if count := tv.ContainerType.FieldCount(); i >= count {
		return nil, fmt.Errorf("cannot get item at field index %d, container only has %d fields", i, count)
	}
	v, err := tv.SubtreeView.Get(i)
	if err != nil {
		return nil, err
	}
	return tv.ContainerType.Fields[i].Type.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ContainerView) Set(i uint64, v View) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ContainerView) setNode(i uint64, b Node) error {
	if fieldCount := tv.ContainerType.FieldCount(); i >= fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, fieldCount)
	}
	if err := tv.SubtreeView.Set(i, b); err != nil {
		return err
	}
	return tv.Hook.PropagateChangeMaybe(tv.Backing())
}

func (tv *ContainerView) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}
