package bitfields

import (
	"fmt"
	"testing"
)

func TestBitlistLen(t *testing.T) {
	cases := []struct {
		v     []byte
		n     uint64
		valid bool
		ones  uint64
	}{
		{[]byte{0}, 0, false, 0},
		{[]byte{0, 0}, 8, false, 0},
		{[]byte{0, 0, 0}, 16, false, 0},
		{[]byte{0xff, 0, 0, 0}, 24, false, 0},
		{[]byte{1, 2, 3, 0xff, 0}, 32, false, 0},
		{[]byte{1}, 0, true, 0},
		{[]byte{2}, 1, true, 0},
		{[]byte{3}, 1, true, 1},
		{[]byte{0x1a}, 4, true, 2},
		{[]byte{0x2b}, 5, true, 3},
		{[]byte{0xab}, 7, true, 4},
		{[]byte{0, 0x9b}, 8 + 7, true, 4},
		{[]byte{0, 0, 0x9b}, 8 + 8 + 7, true, 4},
		{[]byte{0xff, 0xff, 0x9b}, 8 + 8 + 7, true, 8 + 8 + 4},
		{[]byte{0xff, 0xff, 0x04}, 8 + 8 + 2, true, 8 + 8},
		{[]byte{0, 0, 0, 0, 0, 4}, 5*8 + 2, true, 0},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("v %b (bin) len %d", testCase.v, testCase.n), func(t *testing.T) {
			t.Run("get length", func(t *testing.T) {
				if x := BitlistLen(testCase.v); x != testCase.n {
					t.Errorf("expected bitlist to be of length: %d but got %d", testCase.n, x)
				}
			})
			t.Run(fmt.Sprintf("check valid %v", testCase.valid), func(t *testing.T) {
				if err := BitlistCheck(testCase.v, testCase.n); err != nil && testCase.valid {
					t.Errorf("expected bitlist to be valid but got error: %v", err)
				} else if err == nil && !testCase.valid {
					t.Error("expected bitlist to be invalid but got no error")
				}
			})
			if testCase.valid {
				t.Run("check ones count", func(t *testing.T) {
					if res := BitlistOnesCount(testCase.v); res != testCase.ones {
						t.Errorf("expected %d one bits, but got %d", testCase.ones, res)
					}
				})
			}
		})
	}
}
