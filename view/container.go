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

type ContainerTypeDef struct {
	ComplexTypeBase
	Fields []FieldDef
}

func ContainerType(name string, fields []FieldDef) *ContainerTypeDef {
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
	return &ContainerTypeDef{
		ComplexTypeBase: ComplexTypeBase{
			TypeName: name,
			MinSize: minSize,
			MaxSize: maxSize,
			Size: size,
			IsFixedSize: isFixedSize,
		},
		Fields: fields,
	}
}

func (td *ContainerTypeDef) DefaultNode() Node {
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

func (td *ContainerTypeDef) ViewFromBacking(node Node, hook BackingHook) (View, error) {
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

func (td *ContainerTypeDef) Default(hook BackingHook) View {
	return td.New(hook)
}

func (td *ContainerTypeDef) New(hook BackingHook) *ContainerView {
	v, _ := td.ViewFromBacking(td.DefaultNode(), hook)
	return v.(*ContainerView)
}

func (td *ContainerTypeDef) FieldCount() uint64 {
	return uint64(len(td.Fields))
}

func (td *ContainerTypeDef) Deserialize(r io.Reader, scope uint64) error {
	// TODO
	return nil
}

func (td *ContainerTypeDef) String() string {
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
	*ContainerTypeDef
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
	if count := tv.ContainerTypeDef.FieldCount(); i >= count {
		return nil, fmt.Errorf("cannot get item at field index %d, container only has %d fields", i, count)
	}
	v, err := tv.SubtreeView.GetNode(i)
	if err != nil {
		return nil, err
	}
	return tv.ContainerTypeDef.Fields[i].Type.ViewFromBacking(v, tv.ItemHook(i))
}

func (tv *ContainerView) Set(i uint64, v View) error {
	return tv.setNode(i, v.Backing())
}

func (tv *ContainerView) setNode(i uint64, b Node) error {
	if fieldCount := tv.ContainerTypeDef.FieldCount(); i >= fieldCount {
		return fmt.Errorf("cannot set item at field index %d, container only has %d fields", i, fieldCount)
	}
	return tv.SubtreeView.SetNode(i, b)
}

func (tv *ContainerView) ItemHook(i uint64) BackingHook {
	return func(b Node) error {
		return tv.setNode(i, b)
	}
}
