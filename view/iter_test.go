package view

import (
	"fmt"
	. "github.com/protolambda/ztyp/tree"
	"testing"
)

type iterTestCase struct {
	node     Node
	dOffset  uint8
	length   uint64
	depth    uint8
	expected []Node
}

func TestNodeReadonlyIter(t *testing.T) {
	maxLength := 100
	treeEnd := maxLength * 2
	nodes := make([]Node, treeEnd, treeEnd)
	for i := maxLength; i < treeEnd; i++ {
		nodes[i] = &Root{uint8(i)}
	}
	for i := maxLength - 1; i > 0; i-- {
		nodes[i] = NewPairNode(nodes[i*2], nodes[i*2+1])
	}
	testCases := make([]iterTestCase, 0)
	depth := CoverDepth(uint64(maxLength))

	for d := uint8(0); d < depth; d++ {
		for dOffset := uint8(0); dOffset < (depth - d); dOffset++ {
			top := uint64(1) << dOffset
			bottomStart := uint64(1) << (dOffset + d)
			maxLenAtBottom := uint64(1) << (d - dOffset)
			for l := uint64(0); l <= maxLenAtBottom; l++ {
				testCases = append(testCases, iterTestCase{
					node:     nodes[top],
					dOffset:  dOffset,
					length:   l,
					depth:    d,
					expected: nodes[bottomStart : bottomStart+l],
				})
			}
		}
	}
	hFn := GetHashFn()
	nodes[1].MerkleRoot(hFn)

	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("case_dep_%d_len_%d_off_%d", testCase.depth, testCase.length, testCase.dOffset), func(t *testing.T) {
			iter := nodeReadonlyIter(testCase.node, testCase.length, testCase.depth)
			for i := uint64(0); i < testCase.length; i++ {
				el, ok, err := iter.Next()
				if err != nil {
					t.Fatalf("unexpected err on %d: %v", i, err)
				}
				if !ok {
					t.Fatalf("unexpected stop on %d: %v", i, err)
				}
				if el.MerkleRoot(hFn) != testCase.expected[i].MerkleRoot(hFn) {
					t.Fatalf("expected node is different on %d", i)
				}
			}
			el, ok, err := iter.Next()
			if el != nil || ok || err != nil {
				t.Fatal("expected end")
			}
		})
	}
}
