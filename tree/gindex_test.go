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
