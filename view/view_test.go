package view

import (
	"bufio"
	"bytes"
	"encoding/hex"
	"encoding/json"
	"reflect"
	"strings"
	"testing"
)

func bytes_hash(data: bytes):
    return sha256(data).digest()


var EmptyTestStructType = ContainerType("Container", []FieldDef{})

var SingleFieldTestStructType = ContainerType("SingleFieldTestStruct", []FieldDef{
	{Name: "A", Type: ByteType},
})

var SmallTestStructType = ContainerType("SmallTestStruct", []FieldDef{
	{Name: "A", Type: Uint16Type},
	{Name: "B", Type: Uint16Type},
})

var FixedTestStructType = ContainerType("FixedTestStruct", []FieldDef{
	{Name: "A", Type: Uint8Type},
	{Name: "B", Type: Uint64Type},
	{Name: "C", Type: Uint32Type},
})

var VarTestStructType = ContainerType("VarTestStruct", []FieldDef{
	{Name: "A", Type: Uint16Type},
	{Name: "B", Type: ListType("", Uint16Type, 1024)},
	{Name: "C", Type: Uint8Type},
})

var ComplexTestStruct = ContainerType("ComplexTestStruct", []FieldDef{
	{Name: "A", Type: Uint16Type},
	{Name: "B", Type: ListType("", Uint16Type, 128)},
	{Name: "C", Type: Uint8Type},
	{Name: "D", Type: ListType("", ByteType, 256)},
	{Name: "E", Type: VarTestStructType},
	{Name: "F", Type: VectorType("", FixedTestStructType, 4)},
	{Name: "G", Type: VectorType("", VarTestStructType, 2)},
})

var SigType = BasicVectorType("", ByteType, 96)
var sigBytes = [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff}
var SigView, _ = SigType.Deserialize(bytes.NewReader(sigBytes[:]), 96)


func chunk(v string) string {
	res := [32]byte{}
	data, _ := hex.DecodeString(v)
	copy(res[:], data)
	return hex.EncodeToString(res[:])
}

func h(a string, b string) string {
	aBytes, _ := hex.DecodeString(a)
	bBytes, _ := hex.DecodeString(b)
	data := append(append(make([]byte, 0, 64), aBytes...), bBytes...)
	res := sha256.Sum256(data)
	return hex.EncodeToString(res[:])
}

func merge(a string, branch []string) (out string) {
	out = a
	for _, b := range branch {
		out = h(out, b)
	}
	return
}

func repeat(v string, count int) (out string) {
	for i := 0; i < count; i++ {
		out += v
	}
	return
}

type sszTestCase struct {
	// name of test
	name string
	// any value
	value View
	// hex formatted, no prefix
	hex string
	// hex formatted, no prefix
	root string
	// type
	typ TypeDef
}

// note: expected strings are in little-endian, hence the seemingly out of order bytes.
var testCases []sszTestCase

func init() {
	var zeroHashes = []string{chunk("")}

	for layer := 1; layer < 32; layer++ {
		zeroHashes = append(zeroHashes, h(zeroHashes[layer-1], zeroHashes[layer-1]))
	}

	testCases = []sszTestCase{
		{"bool F", BoolView(false), "00", chunk("00"), BoolType},
		{"bool T", BoolView(true), "01", chunk("01"), BoolType},
    	{"bitlist empty", BitlistType("", 8).New(nil), Bitlist[8](), "01", h(chunk(""), chunk("00"))},
		{"bitvector TTFTFTFF", bitvec8{0x2b}, "2b", chunk("2b"), getTyp((*bitvec8)(nil))},
		{"bitlist TTFTFTFF", bitlist8{0x2b, 0x01}, "2b01", h(chunk("2b"), chunk("08")), getTyp((*bitlist8)(nil))},
		{"bitvector FTFT", bitvec4{0x0a}, "0a", chunk("0a"), getTyp((*bitvec4)(nil))},
		{"bitlist FTFT", bitlist4{0x1a}, "1a", h(chunk("0a"), chunk("04")), getTyp((*bitlist4)(nil))},
		{"bitvector FTF", bitvec3{0x02}, "02", chunk("02"), getTyp((*bitvec3)(nil))},
		{"bitlist FTF", bitlist3{0x0a}, "0a", h(chunk("02"), chunk("03")), getTyp((*bitlist3)(nil))},
		{"bitvector TFTFFFTTFT", bitvec10{0xc5, 0x02}, "c502", chunk("c502"), getTyp((*bitvec10)(nil))},
		{"bitlist TFTFFFTTFT", bitlist10{0xc5, 0x06}, "c506", h(chunk("c502"), chunk("0A")), getTyp((*bitlist10)(nil))},
		{"bitvector TFTFFFTTFTFFFFTT", bitvec16{0xc5, 0xc2}, "c5c2", chunk("c5c2"), getTyp((*bitvec16)(nil))},
		{"bitlist TFTFFFTTFTFFFFTT", bitlist16{0xc5, 0xc2, 0x01}, "c5c201", h(chunk("c5c2"), chunk("10")), getTyp((*bitlist16)(nil))},
		{"long bitvector", bitvec512{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
		}, repeat("ff", 64), h(repeat("ff", 32), repeat("ff", 32)), getTyp((*bitvec512)(nil)),
		},
		{"long bitlist", bitlist512{7}, "07", h(h(chunk("03"), chunk("")), chunk("02")), getTyp((*bitlist512)(nil))},
		{"long bitlist filled", bitlist512{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x01,
		}, repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), chunk("0002")), getTyp((*bitlist512)(nil))},
		{"odd bitvector filled", bitvec513{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x01,
		}, repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk(""))), getTyp((*bitvec513)(nil))},
		{"odd bitlist filled", bitlist513{
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff,
			0x03,
		}, repeat("ff", 64) + "03", h(h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk(""))), chunk("0102")), getTyp((*bitlist513)(nil))},
		{"uint8 00", uint8(0x00), "00", chunk("00"), getTyp((*uint8)(nil))},
		{"uint8 01", uint8(0x01), "01", chunk("01"), getTyp((*uint8)(nil))},
		{"uint8 ab", uint8(0xab), "ab", chunk("ab"), getTyp((*uint8)(nil))},

		{"byte 00", uint8(0x00), "00", chunk("00"), getTyp((*uint8)(nil))},
		{"byte 01", uint8(0x01), "01", chunk("01"), getTyp((*uint8)(nil))},
		{"byte ab", uint8(0xab), "ab", chunk("ab"), getTyp((*uint8)(nil))},

		{"uint16 0000", uint16(0x0000), "0000", chunk("0000"), getTyp((*uint16)(nil))},
		{"uint16 abcd", uint16(0xabcd), "cdab", chunk("cdab"), getTyp((*uint16)(nil))},
		{"uint32 00000000", uint32(0x00000000), "00000000", chunk("00000000"), getTyp((*uint32)(nil))},
		{"uint32 01234567", uint32(0x01234567), "67452301", chunk("67452301"), getTyp((*uint32)(nil))},
		{"small {4567, 0123}", smallTestStruct{0x4567, 0x0123}, "67452301", h(chunk("6745"), chunk("2301")), getTyp((*smallTestStruct)(nil))},
		{"small [4567, 0123]::2", [2]uint16{0x4567, 0x0123}, "67452301", chunk("67452301"), getTyp((*[2]uint16)(nil))},
		{"uint32 01234567", uint32(0x01234567), "67452301", chunk("67452301"), getTyp((*uint32)(nil))},
		{"uint64 0000000000000000", uint64(0x00000000), "0000000000000000", chunk("0000000000000000"), getTyp((*uint64)(nil))},
		{"uint64 0123456789abcdef", uint64(0x0123456789abcdef), "efcdab8967452301", chunk("efcdab8967452301"), getTyp((*uint64)(nil))},
		("bytes48", Vector[byte, 48], Vector[byte, 48](*range(48)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f",
     h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f00000000000000000000000000000000")),
    ("raw bytes48", ByteVector[48], ByteVector[48](*range(48)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f",
     h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f00000000000000000000000000000000")),
    ("small empty bytelist", List[byte, 10], List[byte, 10](), "", h(chunk(""), chunk("00"))),
    ("big empty bytelist", List[byte, 2048], List[byte, 2048](), "", h(zero_hashes[6], chunk("00"))),
    ("raw small empty bytelist", ByteList[10], ByteList[10](), "", h(chunk(""), chunk("00"))),
    ("raw big empty bytelist", ByteList[2048], ByteList[2048](), "", h(zero_hashes[6], chunk("00"))),
    ("bytelist 7", List[byte, 7], List[byte, 7](*range(7)), "00010203040506",
     h(chunk("00010203040506"), chunk("07"))),
    ("raw bytelist 7", ByteList[7], ByteList[7](*range(7)), "00010203040506",
     h(chunk("00010203040506"), chunk("07"))),
    ("bytelist 50", List[byte, 50], List[byte, 50](*range(50)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031",
     h(h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f30310000000000000000000000000000"), chunk("32"))),
    ("raw bytelist 50", ByteList[50], ByteList[50](*range(50)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031",
     h(h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f30310000000000000000000000000000"), chunk("32"))),
    ("bytelist 6/256", List[byte, 256], List[byte, 256](*range(6)), "000102030405",
     h(h(h(h(chunk("000102030405"), zero_hashes[0]), zero_hashes[1]), zero_hashes[2]), chunk("06"))),
    ("raw bytelist 6/256", ByteList[256], List[byte, 256](*range(6)), "000102030405",
     h(h(h(h(chunk("000102030405"), zero_hashes[0]), zero_hashes[1]), zero_hashes[2]), chunk("06"))),
    ("sig", Vector[byte, 96], Vector[byte, 96](*sig_test_data),
     "0100000000000000000000000000000000000000000000000000000000000000"
     "0200000000000000000000000000000000000000000000000000000000000000"
     "03000000000000000000000000000000000000000000000000000000000000ff",
     h(h(chunk("01"), chunk("02")),
       h("03000000000000000000000000000000000000000000000000000000000000ff", chunk("")))),
	{"sig", [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff},
	"01" + repeat("00", 31) + "02" + repeat("00", 31) + "03" + repeat("00", 30) + "ff",
	h(h(chunk("01"), chunk("02")), h("03"+repeat("00", 30)+"ff", chunk(""))), getTyp((*[96]byte)(nil))},
    ("raw sig", ByteVector[96], ByteVector[96](*sig_test_data),
     "0100000000000000000000000000000000000000000000000000000000000000"
     "0200000000000000000000000000000000000000000000000000000000000000"
     "03000000000000000000000000000000000000000000000000000000000000ff",
     h(h(chunk("01"), chunk("02")),
       h("03000000000000000000000000000000000000000000000000000000000000ff", chunk("")))),
    ("3 sigs", Vector[ByteVector[96], 3], Vector[ByteVector[96], 3](
        [1] + [0 for i in range(95)],
        [2] + [0 for i in range(95)],
        [3] + [0 for i in range(95)]
    ),
     "01" + ("00" * 95) + "02" + ("00" * 95) + "03" + ("00" * 95),
     h(h(h(h(chunk("01"), chunk("")), zero_hashes[1]), h(h(chunk("02"), chunk("")), zero_hashes[1])),
       h(h(h(chunk("03"), chunk("")), zero_hashes[1]), chunk("")))),
		{"singleFieldTestStruct", singleFieldTestStruct{0xab}, "ab", chunk("ab"), getTyp((*singleFieldTestStruct)(nil))},

		{"uint16 list", list32uint16{0xaabb, 0xc0ad, 0xeeff}, "bbaaadc0ffee",
			h(h(chunk("bbaaadc0ffee"), chunk("")), chunk("03000000")), // max length: 32 * 2 = 64 bytes = 2 chunks
			getTyp((*list32uint16)(nil)),
		},
		{"uint32 list", list128uint32{0xaabb, 0xc0ad, 0xeeff}, "bbaa0000adc00000ffee0000",
			// max length: 128 * 4 = 512 bytes = 16 chunks
			h(merge(chunk("bbaa0000adc00000ffee0000"), zeroHashes[0:4]), chunk("03000000")),
			getTyp((*list128uint32)(nil)),
		},
		{"bytes32 list", list64bytes32{[32]byte{0xbb, 0xaa}, [32]byte{0xad, 0xc0}, [32]byte{0xff, 0xee}},
			"bbaa000000000000000000000000000000000000000000000000000000000000" +
				"adc0000000000000000000000000000000000000000000000000000000000000" +
				"ffee000000000000000000000000000000000000000000000000000000000000",
			h(merge(h(h(chunk("bbaa"), chunk("adc0")), h(chunk("ffee"), chunk(""))), zeroHashes[2:6]), chunk("03000000")),
			getTyp((*list64bytes32)(nil)),
		},
		{"bytes32 list long", list128bytes32{
			{1}, {2}, {3}, {4}, {5}, {6}, {7}, {8}, {9}, {10},
			{11}, {12}, {13}, {14}, {15}, {16}, {17}, {18}, {19},
		},
			"01" + repeat("00", 31) + "02" + repeat("00", 31) +
				"03" + repeat("00", 31) + "04" + repeat("00", 31) +
				"05" + repeat("00", 31) + "06" + repeat("00", 31) +
				"07" + repeat("00", 31) + "08" + repeat("00", 31) +
				"09" + repeat("00", 31) + "0a" + repeat("00", 31) +
				"0b" + repeat("00", 31) + "0c" + repeat("00", 31) +
				"0d" + repeat("00", 31) + "0e" + repeat("00", 31) +
				"0f" + repeat("00", 31) + "10" + repeat("00", 31) +
				"11" + repeat("00", 31) + "12" + repeat("00", 31) +
				"13" + repeat("00", 31),
			h(merge(
				h(
					h(
						h(
							h(h(chunk("01"), chunk("02")), h(chunk("03"), chunk("04"))),
							h(h(chunk("05"), chunk("06")), h(chunk("07"), chunk("08"))),
						),
						h(
							h(h(chunk("09"), chunk("0a")), h(chunk("0b"), chunk("0c"))),
							h(h(chunk("0d"), chunk("0e")), h(chunk("0f"), chunk("10"))),
						),
					),
					h(
						h(
							h(h(chunk("11"), chunk("12")), h(chunk("13"), chunk(""))),
							zeroHashes[2],
						),
						zeroHashes[3],
					),
				),
				// 128 chunks = 7 deep
				zeroHashes[5:7]), chunk("13000000")),
			getTyp((*list128bytes32)(nil)),
		},
		{"fixedTestStruct", fixedTestStruct{A: 0xab, B: 0xaabbccdd00112233, C: 0x12345678}, "ab33221100ddccbbaa78563412",
			h(h(chunk("ab"), chunk("33221100ddccbbaa")), h(chunk("78563412"), chunk(""))), getTyp((*fixedTestStruct)(nil))},
		{"VarTestStruct nil", VarTestStruct{A: 0xabcd, B: nil, C: 0xff}, "cdab07000000ff",
			// log2(1024*2/32)= 6 deep
			h(h(chunk("cdab"), h(zeroHashes[6], chunk("00000000"))), h(chunk("ff"), chunk(""))), getTyp((*VarTestStruct)(nil))},
		{"VarTestStruct empty", VarTestStruct{A: 0xabcd, B: make([]uint16, 0), C: 0xff}, "cdab07000000ff",
			h(h(chunk("cdab"), h(zeroHashes[6], chunk("00000000"))), h(chunk("ff"), chunk(""))), getTyp((*VarTestStruct)(nil))},
		{"VarTestStruct some", VarTestStruct{A: 0xabcd, B: []uint16{1, 2, 3}, C: 0xff}, "cdab07000000ff010002000300",
			h(
				h(
					chunk("cdab"),
					h(
						merge(
							chunk("010002000300"),
							zeroHashes[0:6],
						),
						chunk("03000000"), // length mix in
					),
				),
				h(chunk("ff"), chunk("")),
			),
			getTyp((*VarTestStruct)(nil))},
		{"empty list", ListA{}, "", h(zeroHashes[2], chunk("00000000")), getTyp((*ListA)(nil))},
		{"empty var element list", ListB{}, "", h(zeroHashes[3], chunk("00000000")), getTyp((*ListB)(nil))},
		{"var element list", ListB{
			{A: 0xdead, B: []uint16{1, 2, 3}, C: 0x11},
			{A: 0xbeef, B: []uint16{4, 5, 6}, C: 0x22}},
			"08000000" + "15000000" +
				"adde0700000011010002000300" +
				"efbe0700000022040005000600",
			h(h(
				h(
					h(
						h(h(chunk("adde"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
							h(chunk("11"), chunk(""))),
						h(h(chunk("efbe"), h(merge(chunk("040005000600"), zeroHashes[0:6]), chunk("03000000"))),
							h(chunk("22"), chunk(""))),
					),
					zeroHashes[1],
				),
				zeroHashes[2],
			), chunk("02000000")), getTyp((*ListB)(nil))},
		{"empty list fields", ListStruct{}, "08000000" + "08000000",
			h(h(zeroHashes[2], chunk("")), h(zeroHashes[3], chunk(""))), getTyp((*ListStruct)(nil))},
		{"empty last field", ListStruct{A: ListA{
			smallTestStruct{A: 0xaa11, B: 0xbb22},
			smallTestStruct{A: 0xcc33, B: 0xdd44},
			smallTestStruct{A: 0x1234, B: 0x4567},
		}}, "08000000" + "14000000" + ("11aa22bb" + "33cc44dd" + "34126745"),
			h(
				h(
					h(
						h(
							h(chunk("11aa"), chunk("22bb")),
							h(chunk("33cc"), chunk("44dd")),
						),
						h(
							h(chunk("3412"), chunk("6745")),
							chunk(""),
						),
					),
					chunk("03000000"),
				),
				h(zeroHashes[3], chunk("")),
			), getTyp((*ListStruct)(nil))},
		{"complexTestStruct",
			complexTestStruct{
				A: 0xaabb,
				B: uint16List128{0x1122, 0x3344},
				C: 0xff,
				D: bytelist256("foobar"),
				E: VarTestStruct{A: 0xabcd, B: uint16List1024{1, 2, 3}, C: 0xff},
				F: [4]fixedTestStruct{
					{0xcc, 0x4242424242424242, 0x13371337},
					{0xdd, 0x3333333333333333, 0xabcdabcd},
					{0xee, 0x4444444444444444, 0x00112233},
					{0xff, 0x5555555555555555, 0x44556677}},
				G: [2]VarTestStruct{
					{A: 0xdead, B: []uint16{1, 2, 3}, C: 0x11},
					{A: 0xbeef, B: []uint16{4, 5, 6}, C: 0x22}},
			},
			"bbaa" +
				"47000000" + // offset of B, []uint16
				"ff" +
				"4b000000" + // offset of foobar
				"51000000" + // offset of E
				"cc424242424242424237133713" +
				"dd3333333333333333cdabcdab" +
				"ee444444444444444433221100" +
				"ff555555555555555577665544" +
				"5e000000" + // pointer to G
				"22114433" + // contents of B
				"666f6f626172" + // foobar
				"cdab07000000ff010002000300" + // contents of E
				"08000000" + "15000000" + // [start G]: local offsets of [2]VarTestStruct
				"adde0700000011010002000300" +
				"efbe0700000022040005000600",
			h(
				h(
					h( // A and B
						chunk("bbaa"),
						h(merge(chunk("22114433"), zeroHashes[0:3]), chunk("02000000")), // 2*128/32 = 8 chunks
					),
					h( // C and D
						chunk("ff"),
						h(merge(chunk("666f6f626172"), zeroHashes[0:3]), chunk("06000000")), // 256/32 = 8 chunks
					),
				),
				h(
					h( // E and F
						h(h(chunk("cdab"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
							h(chunk("ff"), chunk(""))),
						h(
							h(
								h(h(chunk("cc"), chunk("4242424242424242")), h(chunk("37133713"), chunk(""))),
								h(h(chunk("dd"), chunk("3333333333333333")), h(chunk("cdabcdab"), chunk(""))),
							),
							h(
								h(h(chunk("ee"), chunk("4444444444444444")), h(chunk("33221100"), chunk(""))),
								h(h(chunk("ff"), chunk("5555555555555555")), h(chunk("77665544"), chunk(""))),
							),
						),
					),
					h( // G and padding
						h(
							h(h(chunk("adde"), h(merge(chunk("010002000300"), zeroHashes[0:6]), chunk("03000000"))),
								h(chunk("11"), chunk(""))),
							h(h(chunk("efbe"), h(merge(chunk("040005000600"), zeroHashes[0:6]), chunk("03000000"))),
								h(chunk("22"), chunk(""))),
						),
						chunk(""),
					),
				),
			),
			getTyp((*complexTestStruct)(nil))},
	}
}


func TestEncode(t *testing.T) {
	var buf bytes.Buffer
	bufWriter := bufio.NewWriter(&buf)

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			if n, err := Encode(bufWriter, tt.value, sszTyp); err != nil {
				t.Fatalf("encoding failed, wrote to byte %d, err: %v", n, err)
			}
			if err := bufWriter.Flush(); err != nil {
				t.Fatal(err)
			}
			data := buf.Bytes()
			if res := fmt.Sprintf("%x", data); res != tt.hex {
				t.Fatalf("encoded different data:\n     got %s\nexpected %s", res, tt.hex)
			}
			if size := SizeOf(tt.value, sszTyp); uint64(len(data)) != size {
				t.Errorf("encoded output does not match expected size:"+
					" len(data): %d but expected: %d", len(data), size)
			}
		})
	}
}

func TestDecode(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(data)
			// For dynamic types, we need to pass the length of the message to the decoder.
			// See SSZ-envelope discussion
			bytesLen := uint64(len(tt.hex)) / 2

			destination := reflect.New(tt.typ).Interface()
			if err := Decode(r, bytesLen, destination, sszTyp); err != nil {
				t.Fatal(err)
			}
			res, err := json.Marshal(destination)
			if err != nil {
				t.Fatal(err)
			}
			expected, err := json.Marshal(tt.value)
			if err != nil {
				t.Fatal(err)
			}
			// adjust expected json string. No effective difference between null and an empty slice. We prefer nil.
			if adjusted := strings.ReplaceAll(string(expected), "[]", "null"); string(res) != adjusted {
				t.Fatalf("decoded different data:\n     got %s\nexpected %s", res, adjusted)
			}
		})
	}
}

func TestHashTreeRoot(t *testing.T) {
	var buf bytes.Buffer

	// re-use a hash function
	sha := sha256.New()
	hashFn := func(input []byte) (out [32]byte) {
		sha.Reset()
		sha.Write(input)
		copy(out[:], sha.Sum(nil))
		return
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			root := HashTreeRoot(hashFn, tt.value, sszTyp)
			res := hex.EncodeToString(root[:])
			if res != tt.root {
				t.Errorf("Expected root %s but got %s", tt.root, res)
			}
		})
	}
}

func TestDryCheck(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sszTyp, err := SSZFactory(tt.typ)
			if err != nil {
				t.Fatal(err)
			}
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(data)
			// For dynamic types, we need to pass the length of the message to the decoder.
			// See SSZ-envelope discussion
			bytesLen := uint64(len(tt.hex)) / 2

			if err := DryCheck(r, bytesLen, sszTyp); err != nil {
				t.Error(err)
			}
		})
	}
}

@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_type_bounds(name: str, typ: Type[View], value: View, serialized: str, root: str):
    byte_len = len(bytes.fromhex(serialized))
    assert typ.min_byte_length() <= byte_len <= typ.max_byte_length()
    if typ.is_fixed_byte_length():
        assert byte_len == typ.type_byte_length()


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_value_byte_length(name: str, typ: Type[View], value: View, serialized: str, root: str):
    assert value.value_byte_length() == len(bytes.fromhex(serialized))


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_typedef(name: str, typ: Type[View], value: View, serialized: str, root: str):
    assert issubclass(typ, TypeDef)


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_serialize(name: str, typ: Type[View], value: View, serialized: str, root: str):
    stream = io.BytesIO()
    length = value.serialize(stream)
    stream.seek(0)
    encoded = stream.read()
    assert encoded.hex() == serialized
    assert length*2 == len(serialized)


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_encode_bytes(name: str, typ: Type[View], value: View, serialized: str, root: str):
    encoded = value.encode_bytes()
    assert encoded.hex() == serialized


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_hash_tree_root(name: str, typ: Type[View], value: View, serialized: str, root: str):
    assert value.hash_tree_root().hex() == root


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_deserialize(name: str, typ: Type[View], value: View, serialized: str, root: str):
    stream = io.BytesIO()
    bytez = bytes.fromhex(serialized)
    stream.write(bytez)
    stream.seek(0)
    decoded = typ.deserialize(stream, len(bytez))
    assert decoded == value


@pytest.mark.parametrize("name, typ, value, serialized, root", test_data)
func test_decode_bytes(name: str, typ: Type[View], value: View, serialized: str, root: str):
    bytez = bytes.fromhex(serialized)
    decoded = typ.decode_bytes(bytez)
    assert decoded == value
