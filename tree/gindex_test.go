package tree

import (
	"encoding/hex"
	"testing"
)

func TestGindex64_Encoding(t *testing.T) {
	cases := map[Gindex64]string{
		0:                  "",
		1:                  "01",
		2:                  "02",
		0xff:               "ff",
		0x01ff:             "01ff",
		0xaaaaaa:           "aaaaaa",
		0xabcdef12345:      "0abcdef12345",
		0xaffffffffffffffb: "affffffffffffffb",
	}
	for k, v := range cases {
		t.Run(v, func(t *testing.T) {
			t.Run("little", func(t *testing.T) {
				data := k.LittleEndian()
				got := hex.EncodeToString(data)

				decoded, err := hex.DecodeString(v)
				if err != nil {
					t.Fatal(err)
				}
				for i, j := 0, len(decoded)-1; i < j; i, j = i+1, j-1 {
					decoded[i], decoded[j] = decoded[j], decoded[i]
				}
				littleEndianV := hex.EncodeToString(decoded)
				if got != littleEndianV {
					t.Errorf("got %s, expected %s", got, littleEndianV)
				}
			})
			t.Run("big", func(t *testing.T) {
				data := k.BigEndian()
				got := hex.EncodeToString(data)
				if got != v {
					t.Errorf("got %s, expected %s", got, v)
				}
			})
		})
	}
}

func TestToGindex64(t *testing.T) {
	cases := []struct {
		index         uint64
		limit         uint64
		expectedDepth uint8
		expectedGi    Gindex64
	}{
		{11, 11, 4, 27},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			depth := CoverDepth(c.limit)
			if depth != c.expectedDepth {
				t.Errorf("got %d, expected %d", depth, c.expectedDepth)
			}
			gi, err := ToGindex64(c.index, depth)
			if err != nil {
				t.Fatal(err)
			}
			if gi != c.expectedGi {
				t.Errorf("got %d, expected %d", gi, c.expectedGi)
			}
		})
	}
}

func TestGindex64Proof(t *testing.T) {
	cases := []struct {
		gindex1 Gindex64
		gindex2 Gindex64
		isProof bool
	}{
		{8, 15, false},
		{8, 14, false},
		{8, 13, false},
		{8, 12, false},
		{8, 11, false},
		{8, 10, false},
		{8, 9, true},
		{8, 8, false},
		{8, 7, false},
		{8, 6, false},
		{8, 5, true},
		{8, 4, false},
		{8, 3, true},
		{8, 2, false},
		{8, 1, true},
	}
	for _, c := range cases {
		if c.gindex2.IsProof(c.gindex1) != c.isProof {
			t.Errorf("got %v, expected %v", c.gindex2.IsProof(c.gindex1), c.isProof)
		}
	}
}

func TestGindex64Split(t *testing.T) {
	cases := []struct {
		gindex          Gindex64
		depth           uint32
		expectedGindex1 Gindex64
		expectedGindex2 Gindex64
	}{
		{221184, 0, 1, 221184},
		{221184, 4, 27, 8192},
		{212992, 4, 26, 8192},
		{221183, 4, 26, 16383},
		{27, 4, 27, 1},
	}
	for _, c := range cases {
		t.Run("", func(t *testing.T) {
			idx1, idx2 := c.gindex.Split(c.depth)
			if idx1 != c.expectedGindex1 {
				t.Errorf("got %d, expected %d", idx1, c.expectedGindex1)
			}
			if idx2 != c.expectedGindex2 {
				t.Errorf("got %d, expected %d", idx2, c.expectedGindex2)
			}
		})
	}
}
