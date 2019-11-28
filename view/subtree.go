package view

import (
	. "github.com/protolambda/ztyp/tree"
)

type SubtreeView struct {
	BackingNode Node
	depth       uint8
}

func (stv *SubtreeView) Get(i uint64) (Node, error) {
	return stv.BackingNode.Getter(i, stv.depth)
}

func (stv *SubtreeView) Set(i uint64, node Node) error {
	s, err := stv.BackingNode.Setter(i, stv.depth)
	if err != nil {
		return err
	}
	stv.BackingNode = s(node)
	return nil
}

func (stv *SubtreeView) Backing() Node {
	return stv.BackingNode
}
