package view

import "testing"

func TestRootView_MarshalText(t *testing.T) {
	r := RootView{0: 0xab, 31: 0x12}
	s, err := r.MarshalText()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
	if string(s) != "0xab00000000000000000000000000000000000000000000000000000000000012" {
		t.Errorf("unexpected string: %s", string(s))
	}
}

func TestRootView_UnmarshalText(t *testing.T) {
	cases := []struct {
		r RootView
		s string
	}{
		{RootView{0: 0xab, 31: 0x12}, "ab00000000000000000000000000000000000000000000000000000000000012"},
		{RootView{0: 0xab, 31: 0x12}, "0Xab00000000000000000000000000000000000000000000000000000000000012"},
		{RootView{0: 0xab, 31: 0x12}, "0xab00000000000000000000000000000000000000000000000000000000000012"},
		{RootView{}, "0000000000000000000000000000000000000000000000000000000000000000"},
		{RootView{}, "0x0000000000000000000000000000000000000000000000000000000000000000"},
	}
	for _, c := range cases {
		t.Run(c.s, func(t *testing.T) {
			var root RootView
			if err := root.UnmarshalText([]byte(c.s)); err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if root != c.r {
				t.Errorf("unexpected root: %x", root[:])
			}
		})
	}
	bad := []string{
		// no quotes. It's marshalText, not raw JSON.
		`""`, `"0x0000000000000000000000000000000000000000000000000000000000000000"`,
		"00000000000000000000000000000000000000000000000000000000000000000", // 1 zero extra
		"00x00000000000000000000000000000000000000000000000000000000000000000",
		"00x0000000000000000000000000000000000000000000000000000000000000000",
		"00000000000000000000000000000000000000000000000000000000000000000x",
		"0x000x00000000000000000000000000000000000000000000000000000000000000",
		"0x0g00000000000000000000000000000000000000000000000000000000000012",
	}
	for _, b := range bad {
		t.Run(b, func(t *testing.T) {
			var root RootView
			if err := root.UnmarshalText([]byte(b)); err == nil {
				t.Errorf("expected error, but got root %x", root[:])
			}
		})
	}
}
