package bitfields

import (
	"fmt"
	"testing"
)

func TestBitIndex(t *testing.T) {
	cases := []struct {
		v     byte
		index uint64
	}{
		{0, 0},
		{1, 0},
		{2, 1},
		{3, 1},
		{4, 2},
		{34, 5},
		{127, 6},
		{128, 7},
		{255, 7},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("v %b (bin) index %d", testCase.v, testCase.index), func(t *testing.T) {
			if res := BitIndex(testCase.v); res != testCase.index {
				t.Errorf("unexpected bit index: %d for value %b (bin), expected index: %d",
					res, testCase.v, testCase.index)
			}
		})
	}
}

func BenchmarkBitIndex(b *testing.B) {
	// sum results for fun, and verify it has the same result with different benched solutions with same N.
	out := uint64(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		out += BitIndex(byte(i))
	}
	b.StopTimer()
	b.Logf("result after %d runs: %d", b.N, out)
}

// For speed performance comparison. 2x faster, but dependent on global 256 bytes var.
// (Also consider heap/stack location, take 2x with a grain of salt)
var lookup = [256]byte{
	0,
	0,
	1, 1,
	2, 2, 2, 2,
	3, 3, 3, 3, 3, 3, 3, 3,
	4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4, 4,
	5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5, 5,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6, 6,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
	7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7, 7,
}

func BenchmarkLookupBitIndex(b *testing.B) {
	// sum results for fun, and verify it has the same result with different benched solutions with same N.
	out := uint64(0)
	b.StartTimer()
	for i := 0; i < b.N; i++ {
		out += uint64(lookup[byte(i)])
	}
	b.StopTimer()
	b.Logf("result after %d runs: %d", b.N, out)
}

func TestSetBit(t *testing.T) {
	cases := []struct {
		v         []byte
		index     uint64
		expected1 string
		expected0 string
	}{
		{[]byte{0}, 0, "00000001 ", "00000000 "},
		{[]byte{1}, 1, "00000011 ", "00000001 "},
		{[]byte{1}, 2, "00000101 ", "00000001 "},
		{[]byte{8}, 1, "00001010 ", "00001000 "},
		{[]byte{10}, 1, "00001010 ", "00001000 "},
		{[]byte{255}, 5, "11111111 ", "11011111 "},
		{[]byte{0x00, 0x00}, 0, "00000001 00000000 ", "00000000 00000000 "},
		{[]byte{0x00, 0x00}, 3, "00001000 00000000 ", "00000000 00000000 "},
		{[]byte{0x00, 0x00}, 15, "00000000 10000000 ", "00000000 00000000 "},
		{[]byte{0xff, 0xff}, 0, "11111111 11111111 ", "11111110 11111111 "},
		{[]byte{0xff, 0xff}, 5, "11111111 11111111 ", "11011111 11111111 "},
		{[]byte{0x13, 0x37}, 5, "00110011 00110111 ", "00010011 00110111 "},
		{[]byte{0x13, 0x37}, 8, "00010011 00110111 ", "00010011 00110110 "},
		{[]byte{0xff, 0xff}, 10, "11111111 11111111 ", "11111111 11111011 "},
		{[]byte{0x13, 0x37}, 10, "00010011 00110111 ", "00010011 00110011 "},
		{[]byte{0xff, 0xff}, 15, "11111111 11111111 ", "11111111 01111111 "},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("v %b (bin) index %d", testCase.v, testCase.index), func(t *testing.T) {
			t.Run("set to 1", func(t *testing.T) {
				a := make([]byte, len(testCase.v))
				copy(a, testCase.v)
				SetBit(a, testCase.index, true)
				res := ""
				for _, b := range a {
					res += fmt.Sprintf("%08b ", b)
				}
				if res != testCase.expected1 {
					t.Errorf("expected %s but got %s", testCase.expected1, res)
				}
			})
			t.Run("set to 0", func(t *testing.T) {
				a := make([]byte, len(testCase.v))
				copy(a, testCase.v)
				SetBit(a, testCase.index, false)
				res := ""
				for _, b := range a {
					res += fmt.Sprintf("%08b ", b)
				}
				if res != testCase.expected0 {
					t.Errorf("expected %s but got %s", testCase.expected1, res)
				}
			})
			t.Run("get bit", func(t *testing.T) {
				data := ""
				for _, b := range testCase.v {
					data += fmt.Sprintf("%08b", b)
				}
				bit := GetBit(testCase.v, testCase.index)
				expected := data[(testCase.index/8)*8+7-(testCase.index%8)] == '1'
				if bit != expected {
					t.Errorf("expected %v but got %v, data: %s", bit, expected, data)
				}
			})
		})
	}
}

func TestIsZeroBitlist(t *testing.T) {
	cases := []struct {
		v      []byte
		isZero bool
	}{
		{[]byte{0}, true},
		{[]byte{1}, true},
		{[]byte{2}, true},
		{[]byte{3}, false},
		{[]byte{5}, false},
		{[]byte{8}, true},
		{[]byte{0x80}, true},
		{[]byte{0x81}, false},
		{[]byte{0x90}, false},
		{[]byte{0, 0}, true},
		{[]byte{0, 1}, true},
		{[]byte{0, 2}, true},
		{[]byte{2, 2}, false},
		{[]byte{0, 3}, false},
		{[]byte{0, 5}, false},
		{[]byte{0, 8}, true},
		{[]byte{0, 0x80}, true},
		{[]byte{3, 0x80}, false},
		{[]byte{0, 0x81}, false},
		{[]byte{0, 0x90}, false},
		{[]byte{0, 0, 0}, true},
		{[]byte{0, 0, 1}, true},
		{[]byte{0, 0, 2}, true},
		{[]byte{0, 0, 3}, false},
		{[]byte{0, 0, 5}, false},
		{[]byte{0, 0, 8}, true},
		{[]byte{0, 0, 0x80}, true},
		{[]byte{0, 0, 0x81}, false},
		{[]byte{0, 0, 0x90}, false},
		{[]byte{3, 0, 0x80}, false},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("is-zero checking %b (bin)", testCase.v), func(t *testing.T) {
			if res := IsZeroBitlist(testCase.v); res != testCase.isZero {
				t.Errorf("unexpected is-zero result for %b (bin). Expected: %v, got: %v",
					testCase.v, testCase.isZero, res)
			}
		})
	}
}

func TestCovers(t *testing.T) {
	cases := []struct {
		a      []byte
		b      []byte
		covers bool
		err    bool
	}{
		{[]byte{0}, []byte{}, false, true},
		{[]byte{}, []byte{0}, false, true},
		{[]byte{0}, []byte{1}, false, false},
		{[]byte{1}, []byte{0}, true, false},
		{[]byte{0xff}, []byte{0}, true, false},
		{[]byte{2}, []byte{0xff}, false, false},
		{[]byte{0xf0}, []byte{0x10}, true, false},
		{[]byte{0xf0}, []byte{0x0f}, false, false},
		{[]byte{0xff, 0xff}, []byte{0, 0}, true, false},
		{[]byte{0xff, 0xf0}, []byte{0, 0}, true, false},
		{[]byte{0xff, 0xf0}, []byte{0, 3}, false, false},
		{[]byte{0xff, 0xf0}, []byte{1, 3}, false, false},
		{[]byte{0xff, 0xff, 0xff}, []byte{0xff, 0xff}, false, true},
		{[]byte{0xff, 0xff, 0xff}, []byte{0xff, 0x7f, 0xff}, true, false},
	}
	for _, testCase := range cases {
		t.Run(fmt.Sprintf("covers check %b <> %b (%v)", testCase.a, testCase.b, testCase.covers), func(t *testing.T) {
			if res, err := Covers(testCase.a, testCase.b); (err != nil) != testCase.err {
				if err == nil {
					t.Errorf("expected covers error result between %b <> %b", testCase.a, testCase.b)
				} else {
					t.Errorf("expected covers error %q between %b <> %b", err.Error(), testCase.a, testCase.b)
				}
			} else if res != testCase.covers {
				t.Errorf("unexpected covers result between %b <> %b. Expected: %v, got: %v",
					testCase.a, testCase.b, testCase.covers, res)
			}
		})
	}
}
