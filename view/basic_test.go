package view

import "testing"

func TestUint64View_MarshalJSON(t *testing.T) {
	cases := []struct {
		n uint64
		s string
	}{
		{0, "0"},
		{1, "1"},
		{1234, "1234"},
		{uint64(1) << 63, "9223372036854775808"},
		{^uint64(0), "18446744073709551615"},
	}
	for _, c := range cases {
		t.Run(c.s, func(t *testing.T) {
			v, err := Uint64View(c.n).MarshalJSON()
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if string(v) != `"`+c.s+`"` {
				t.Errorf("unexpected value: %s", string(v))
			}
		})
	}
}

func TestUint64View_UnmarshalJSON(t *testing.T) {
	cases := []struct {
		n uint64
		s string
	}{
		{0, "0"},
		{0, "0b0"},
		{0, "0x0"},
		{0, "0x00"},
		{0, "0x0000"},
		{0, "000"},
		{1, "1"},
		{1234, "1234"},
		{0xc0fFeE, "0xc0fFeE"},
		{0b10101011, "0b10101011"},
		{uint64(1) << 63, "9223372036854775808"},
		{^uint64(0), "18446744073709551615"},
		{^uint64(0), "0xFFFFFFFFFFFFFFFF"},
		{^uint64(0), "0x0FFFFFFFFFFFFFFFF"},
		{^uint64(0), "0x00FFFFFFFFFFFFFFFF"},
	}
	for _, c := range cases {
		t.Run(c.s+"_unquoted", func(t *testing.T) {
			expected := Uint64View(c.n)
			var res Uint64View
			err := res.UnmarshalJSON([]byte(c.s))
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if res != expected {
				t.Errorf("unexpected value: %v", res)
			}
		})
		t.Run(c.s+"_double_quoted", func(t *testing.T) {
			expected := Uint64View(c.n)
			var res Uint64View
			b := []byte{'"'}
			b = append(b, c.s...)
			b = append(b, '"')
			err := res.UnmarshalJSON(b)
			if err != nil {
				t.Errorf("unexpected error: %v", err)
			}
			if res != expected {
				t.Errorf("unexpected value: %v", res)
			}
		})
	}
	bad := []string{
		``, `"`, `""`, `''`, `""0""`, `"0""`, `""0"`, `"0'`, `'0'"`, `""0''`,
		"00x0", "00x0", "0x", "0b", "-0", "-123", "0x1FFFFFFFFFFFFFFFF", "FFFFFFFFFFFFFFFF", "-a", "-0xab",
	}
	for _, b := range bad {
		t.Run(b, func(t *testing.T) {
			var res Uint64View
			if err := res.UnmarshalJSON([]byte(b)); err == nil {
				t.Errorf("expected error, but got value %d", uint64(res))
			}
		})
	}
}
