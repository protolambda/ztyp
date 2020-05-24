package tree

import (
	"fmt"
	"testing"
)

type testCase struct {
	v uint64
	d uint8
	l uint8
	i uint8
}

var testCases = []testCase{
	{v: 0, d: 0, l: 0, i: 0}, // 0
	{v: 1, d: 0, l: 1, i: 0}, // 1
	{v: 2, d: 1, l: 2, i: 1}, // 10
	{v: 3, d: 2, l: 2, i: 1}, // 11
	{v: 4, d: 2, l: 3, i: 2}, // 100
	{v: 5, d: 3, l: 3, i: 2}, // 101
	{v: 6, d: 3, l: 3, i: 2}, // 110
	{v: 7, d: 3, l: 3, i: 2}, // 111
	{v: 8, d: 3, l: 4, i: 3}, // 1000
	{v: 9, d: 4, l: 4, i: 3}, // 1001
	{v: ^uint64(0), d: 64, l: 64, i: 63},
}

func init() {
	for i := uint8(4); i < 64; i++ {
		testCases = append(testCases,
			testCase{v: (1 << i) - 1, d: i, l: i, i: i - 1},
			testCase{v: 1 << i, d: i, l: i + 1, i: i},
			testCase{v: (1 << i) + 1, d: i + 1, l: i + 1, i: i},
		)
	}
}

func TestCoverDepth(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%d_%d", testCase.v, testCase.d), func(t *testing.T) {
			d := CoverDepth(testCase.v)
			if d != testCase.d {
				t.Errorf("Expected depth %d for v %d (bin %b) but got depth %d", testCase.d, testCase.v, testCase.v, d)
			}
		})
	}
}

func TestBitLength(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%d_%d", testCase.v, testCase.l), func(t *testing.T) {
			l := BitLength(testCase.v)
			if l != testCase.l {
				t.Errorf("Expected length %d for v %d (bin %b) but got length %d", testCase.l, testCase.v, testCase.v, l)
			}
		})
	}
}

func TestBitIndex(t *testing.T) {
	for _, testCase := range testCases {
		t.Run(fmt.Sprintf("case_%d_%d", testCase.v, testCase.i), func(t *testing.T) {
			i := BitIndex(testCase.v)
			if i != testCase.i {
				t.Errorf("Expected index %d for v %d (bin %b) but got index %d", testCase.i, testCase.v, testCase.v, i)
			}
		})
	}
}
