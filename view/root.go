package view

import (
	"fmt"
	"github.com/protolambda/ztyp/tree"
)

type RootMeta uint8

func (RootMeta) DefaultNode() tree.Node {
	return &tree.ZeroHashes[0]
}

func (RootMeta) ViewFromBacking(node tree.Node, _ ViewHook) (View, error) {
	root, ok := node.(*tree.Root)
	if !ok {
		return nil, fmt.Errorf("node is not a root: %v", node)
	} else {
		return root, nil
	}
}

const RootType RootMeta = 0
