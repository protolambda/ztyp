package bitfields

import (
	"fmt"
	"testing"
)

func TestBitvectorCheck(t *testing.T) {
	cases := []struct {
		v     []byte
		n     uint64
		valid bool
		ones  uint64
	}{
		{[]byte{}, 0, true, 0},
		{[]byte{0}, 0, false, 0},
		{[]byte{0}, 1, true, 0},
		{[]byte{0}, 8, true, 0},
		{[]byte{1}, 1, true, 1},
		{[]byte{0x31}, 8, true, 3},
		{[]byte{0x14}, 5, true, 2},
		{[]byte{0}, 9, false, 0},
		{[]byte{0, 0}, 9, true, 0},
		{[]byte{0, 0, 0}, 16, false, 0},
		{[]byte{0, 0, 0}, 17, true, 0},
		{[]byte{0, 0, 0}, 24, true, 0},
		{[]byte{0, 0, 0, 0, 0, 0}, 48, true, 0},
		{[]byte{0xff}, 8, true, 8},
		{[]byte{0xff}, 7, false, 0},
		{[]byte{0x7f}, 7, true, 7},
		{[]byte{0xff, 0x80}, 16, true, 8 + 1},
		{[]byte{0xff, 0x80}, 15, false, 0},
		{[]byte{0xff, 0x40}, 15, true, 8 + 1},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("v %b (bin) len %d", testCase.v, testCase.n), func(t *testing.T) {
			t.Run(fmt.Sprintf("check valid %v", testCase.valid), func(t *testing.T) {
				if err := BitvectorCheck(testCase.v, testCase.n); err != nil && testCase.valid {
					t.Errorf("expected bitvector to be valid but got error: %v", err)
				} else if err == nil && !testCase.valid {
					t.Error("expected bitvector to be invalid but got no error")
				}
			})
			if testCase.valid {
				t.Run(fmt.Sprintf("check ones count %v", testCase.valid), func(t *testing.T) {
					if res := BitvectorOnesCount(testCase.v); res != testCase.ones {
						t.Errorf("expected %d ones, but got %d", testCase.ones, res)
					}
				})
			}
		})
	}
}
