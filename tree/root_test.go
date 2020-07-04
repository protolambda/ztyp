package tree

import (
	"encoding/json"
	"fmt"
	"testing"
)

type rootTestCase struct {
	root  Root
	valid bool
	input []byte
}

var unmarshalTestCases = []rootTestCase{
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("aa000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("0xaa000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("0Xaa000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("0XaA000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("AA000000000000000000000000000000000000000000000000000000000000BB")},
	{Root{0: 0xaa, 31: 0xbb}, true, []byte("0XAA000000000000000000000000000000000000000000000000000000000000BB")},
	{Root{}, false, []byte("aa000000000000000000000000000000000000000000000000000000000000bb0x")},
	{Root{}, false, []byte("0xaa000000000000000000000000000000000000000000000000000000000000bb0")},
	{Root{}, false, []byte("0xaa000000000000000000000000000000000000000000000000000000000000b")},
	{Root{}, false, []byte("xaa000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{}, false, []byte("a000000000000000000000000000000000000000000000000000000000000bb")},
	{Root{}, false, []byte("aa00000000000000000000000000000000000000y000000000000000000000bb")},
	{Root{}, false, []byte("aa00")},
	{Root{}, false, []byte("a")},
	{Root{}, false, []byte{}},
}

func TestRoot_UnmarshalText(t *testing.T) {
	for i, testCase := range unmarshalTestCases {
		t.Run(fmt.Sprintf("case%d", i), func(t *testing.T) {
			var x Root
			err := x.UnmarshalText(testCase.input)
			if testCase.valid {
				if err != nil {
					t.Error(err)
					return
				}
				if x != testCase.root {
					t.Errorf("Expected %x but got %x", testCase.root[:], x[:])
				}
			} else {
				if err == nil {
					t.Error("expected error but did not get any")
				}
			}
		})
	}
}

func TestRoot_MarshalText(t *testing.T) {
	x := Root{0: 0xaa, 31: 0xbb}
	data, err := x.MarshalText()
	if err != nil {
		t.Error(t)
		return
	}
	if string(data) != "0xaa000000000000000000000000000000000000000000000000000000000000bb" {
		t.Errorf("unexpected string: %s", string(data))
		return
	}
}

func TestRoot_String(t *testing.T) {
	x := Root{0: 0xaa, 31: 0xbb}
	if res := fmt.Sprintf("fmt %s", x); res != "fmt 0xaa000000000000000000000000000000000000000000000000000000000000bb" {
		t.Errorf("unexpected direct fmt result: %s", res)
		return
	}
	if res := fmt.Sprintf("fmt %s", &x); res != "fmt 0xaa000000000000000000000000000000000000000000000000000000000000bb" {
		t.Errorf("unexpected pointer fmt result: %s", res)
		return
	}
}

func TestRoot_JSON_Marshal(t *testing.T) {
	x := Root{0: 0xaa, 31: 0xbb}
	out, err := json.Marshal(x)
	if err != nil {
		t.Error(err)
		return
	}
	if string(out) != "\"0xaa000000000000000000000000000000000000000000000000000000000000bb\"" {
		t.Errorf("unexpected json: %s", string(out))
		return
	}
}

func TestRoot_JSON_Unmarshal(t *testing.T) {
	var x Root
	if err := json.Unmarshal([]byte("\"0xaa000000000000000000000000000000000000000000000000000000000000bb\""), &x); err != nil {
		t.Error(err)
		return
	}
	if x != (Root{0: 0xaa, 31: 0xbb}) {
		t.Errorf("unexpected value: %s", x)
		return
	}
}
