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
	}{
		{[]byte{}, 0, true},
		{[]byte{0}, 0, false},
		{[]byte{0}, 1, true},
		{[]byte{0}, 8, true},
		{[]byte{0}, 9, false},
		{[]byte{0, 0}, 9, true},
		{[]byte{0, 0, 0}, 16, false},
		{[]byte{0, 0, 0}, 17, true},
		{[]byte{0, 0, 0}, 24, true},
		{[]byte{0, 0, 0, 0, 0, 0}, 48, true},
		{[]byte{0xff}, 8, true},
		{[]byte{0xff}, 7, false},
		{[]byte{0x7f}, 7, true},
		{[]byte{0xff, 0x80}, 16, true},
		{[]byte{0xff, 0x80}, 15, false},
		{[]byte{0xff, 0x40}, 15, true},
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
		})
	}
}
