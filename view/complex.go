package view

const OffsetByteLength = 4

type ComplexTypeBase struct {
	TypeName string
	MinSize uint64
	MaxSize uint64
	Size uint64
	IsFixedSize bool
}

func (td *ComplexTypeBase) Name() string {
	return td.TypeName
}

func (td *ComplexTypeBase) IsFixedByteLength() bool {
	return td.IsFixedSize
}

func (td *ComplexTypeBase) TypeByteLength() uint64 {
	return td.Size
}

func (td *ComplexTypeBase) MinByteLength() uint64 {
	return td.MinSize
}

func (td *ComplexTypeBase) MaxByteLength() uint64 {
	return td.MaxSize
}

type VectorTypeDef interface {
	ElementType() TypeDef
	Length() uint64
}

func VectorType(name string, elemType TypeDef, length uint64) VectorTypeDef {
	basicElemType, ok := elemType.(BasicTypeDef)
	if ok {
		return BasicVectorType(name, basicElemType, length)
	} else {
		return ComplexVectorType(name, elemType, length)
	}
}

type ListTypeDef interface {
	ElementType() TypeDef
	Limit() uint64
}

func ListType(name string, elemType TypeDef, length uint64) ListTypeDef {
	basicElemType, ok := elemType.(BasicTypeDef)
	if ok {
		return BasicVectorType(name, basicElemType, length)
	} else {
		return ComplexVectorType(name, elemType, length)
	}
}
