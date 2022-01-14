package conv

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"
	"testing"
)

type testCase struct {
	str     string
	num     uint64
	bitSize int
	err     error
}

func (tc *testCase) testUnmarshal(t *testing.T) {
	var err error
	var out uint64
	switch tc.bitSize {
	case 8:
		var y uint8
		err = Uint8Unmarshal(&y, []byte(tc.str))
		out = uint64(y)
	case 16:
		var y uint16
		err = Uint16Unmarshal(&y, []byte(tc.str))
		out = uint64(y)
	case 32:
		var y uint32
		err = Uint32Unmarshal(&y, []byte(tc.str))
		out = uint64(y)
	case 64:
		var y uint64
		err = Uint64Unmarshal(&y, []byte(tc.str))
		out = y
	default:
		panic("bad bit size")
	}
	if tc.err != nil {
		if !errors.Is(err, tc.err) {
			t.Errorf("expected err %v, but got %v", tc.err, err)
		}
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if out != tc.num {
			t.Errorf("Unexpected unmarshal result %d, expected %d", out, tc.num)
		}
	}
}

func (tc *testCase) testMarshal(t *testing.T) {
	var out []byte
	var err error
	switch tc.bitSize {
	case 8:
		out, err = Uint8Marshal(uint8(tc.num))
	case 16:
		out, err = Uint16Marshal(uint16(tc.num))
	case 32:
		out, err = Uint32Marshal(uint32(tc.num))
	case 64:
		out, err = Uint64Marshal(tc.num)
	default:
		panic("bad bit size")
	}

	if tc.err != nil {
		if !errors.Is(err, tc.err) {
			t.Errorf("expected err %v, but got %v", tc.err, err)
		}
	} else {
		if err != nil {
			t.Errorf("Unexpected error: %v", err)
		} else if string(out) != tc.str {
			t.Errorf("Unexpected marshal result %s, expected %s", string(out), tc.str)
		}
	}
}

func TestUintUnmarshal(t *testing.T) {
	var testCases []testCase
	for bitSize := 8; bitSize <= 64; bitSize *= 2 {
		max := (uint64(1) << bitSize) - 1
		outOfRange := new(big.Int).Add(new(big.Int).SetUint64(max), big.NewInt(1))

		for _, form := range []string{"%d", "%#b", "%#o", "%#O", "%#x", "%#X"} {
			testCases = append(testCases,
				testCase{fmt.Sprintf(form, 0), 0, bitSize, nil},
				testCase{fmt.Sprintf(form, 1), 1, bitSize, nil},
				testCase{fmt.Sprintf(form, max/3), max / 3, bitSize, nil},
				testCase{fmt.Sprintf(form, max), max, bitSize, nil},
				testCase{fmt.Sprintf(form, outOfRange), 0, bitSize, strconv.ErrRange})
		}
	}
	for _, tc := range testCases {
		testCases = append(testCases, testCase{"-" + tc.str, 0, tc.bitSize, strconv.ErrSyntax})
	}
	for _, tc := range testCases {
		testCases = append(testCases,
			testCase{"\"" + tc.str + "\"", tc.num, tc.bitSize, tc.err},
			testCase{"\"" + tc.str, 0, tc.bitSize, MissingQuoteErr},
			testCase{"'" + tc.str + "'", tc.num, tc.bitSize, strconv.ErrSyntax})
		if tc.err == nil {
			testCases = append(testCases, testCase{tc.str + "\"", 0, tc.bitSize, strconv.ErrSyntax})
		}
	}
	for bitSize := 8; bitSize <= 64; bitSize *= 2 {
		testCases = append(testCases, testCase{"", 0, bitSize, EmptyInputErr})
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("unmarshal_uint%d_%s", tc.bitSize, tc.str), func(t *testing.T) {
			tc.testUnmarshal(t)
		})
	}
}

func TestUintMarshal(t *testing.T) {
	var testCases []testCase
	for bitSize := 8; bitSize <= 64; bitSize *= 2 {
		max := (uint64(1) << bitSize) - 1
		testCases = append(testCases,
			testCase{fmt.Sprintf("\"%d\"", 0), 0, bitSize, nil},
			testCase{fmt.Sprintf("\"%d\"", 1), 1, bitSize, nil},
			testCase{fmt.Sprintf("\"%d\"", max/3), max / 3, bitSize, nil},
			testCase{fmt.Sprintf("\"%d\"", max), max, bitSize, nil})
	}
	for _, tc := range testCases {
		t.Run(fmt.Sprintf("marshal_uint%d_%d", tc.bitSize, tc.num), func(t *testing.T) {
			tc.testMarshal(t)
		})
	}
}
