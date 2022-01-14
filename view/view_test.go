package view

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/protolambda/ztyp/codec"
	"github.com/protolambda/ztyp/tree"
	"reflect"
	"strings"
	"testing"
)

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
	{Name: "B", Type: ListType(Uint16Type, 1024)},
	{Name: "C", Type: Uint8Type},
})

var ComplexTestStructType = ContainerType("ComplexTestStruct", []FieldDef{
	{Name: "A", Type: Uint16Type},
	{Name: "B", Type: ListType(Uint16Type, 128)},
	{Name: "C", Type: Uint8Type},
	{Name: "D", Type: ListType(ByteType, 256)},
	{Name: "E", Type: VarTestStructType},
	{Name: "F", Type: VectorType(FixedTestStructType, 4)},
	{Name: "G", Type: VectorType(VarTestStructType, 2)},
})

var ListAType = ComplexListType(SmallTestStructType, 4)
var ListBType = ComplexListType(VarTestStructType, 8)

var ListStructType = ContainerType("ComplexTestStruct", []FieldDef{
	{Name: "A", Type: ListAType},
	{Name: "B", Type: ListBType},
})

var SigType = BasicVectorType(ByteType, 96)
var sigBytes = [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff}
var SigView, _ = SigType.Deserialize(codec.NewDecodingReader(bytes.NewReader(sigBytes[:]), 96))

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

func repeat(v string, count int) string {
	var buf strings.Builder
	for i := 0; i < count; i++ {
		buf.WriteString(v)
	}
	return buf.String()
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
}

// note: expected strings are in little-endian, hence the seemingly out of order bytes.
var testCases []sszTestCase

func init() {
	var zeroHashes = []string{chunk("")}

	for layer := 1; layer < 32; layer++ {
		zeroHashes = append(zeroHashes, h(zeroHashes[layer-1], zeroHashes[layer-1]))
	}

	bits := func(s string) []bool {
		out := make([]bool, len(s), len(s))
		for i, c := range s {
			if c == 'T' {
				out[i] = true
			}
		}
		return out
	}

	bitvec := func(s string) *BitVectorView {
		v, err := BitVectorType(uint64(len(s))).FromBits(bits(s))
		if err != nil {
			panic(err)
		}
		return v
	}

	bitlist := func(s string, limit uint64) *BitListView {
		if limit < uint64(len(s)) {
			panic("bitlist limit too small")
		}
		v, err := BitListType(limit).FromBits(bits(s))
		if err != nil {
			panic(err)
		}
		return v
	}

	viewMust := func(v View, err error) View {
		if err != nil {
			panic(err)
		}
		return v
	}

	testCases = []sszTestCase{
		{"bool F", BoolView(false), "00", chunk("00")},
		{"bool T", BoolView(true), "01", chunk("01")},
		{"bitlist empty", BitListType(8).New(), "01", h(chunk(""), chunk("00"))},
		{"bitvector TTFTFTFF", bitvec("TTFTFTFF"), "2b", chunk("2b")},
		{"bitlist TTFTFTFF", bitlist("TTFTFTFF", 8), "2b01", h(chunk("2b"), chunk("08"))},
		{"bitvector FTFT", bitvec("FTFT"), "0a", chunk("0a")},
		{"bitlist FTFT", bitlist("FTFT", 4), "1a", h(chunk("0a"), chunk("04"))},
		{"bitvector FTF", bitvec("FTF"), "02", chunk("02")},
		{"bitlist FTF", bitlist("FTF", 3), "0a", h(chunk("02"), chunk("03"))},
		{"bitvector TFTFFFTTFT", bitvec("TFTFFFTTFT"), "c502", chunk("c502")},
		{"bitlist TFTFFFTTFT", bitlist("TFTFFFTTFT", 10), "c506", h(chunk("c502"), chunk("0A"))},
		{"bitvector TFTFFFTTFTFFFFTT", bitvec("TFTFFFTTFTFFFFTT"), "c5c2", chunk("c5c2")},
		{"bitlist TFTFFFTTFTFFFFTT", bitlist("TFTFFFTTFTFFFFTT", 16), "c5c201", h(chunk("c5c2"), chunk("10"))},
		{"long bitvector", bitvec(repeat("T", 512)), repeat("ff", 64), h(repeat("ff", 32), repeat("ff", 32))},
		{"long bitlist", bitlist("TT", 512), "07", h(h(chunk("03"), chunk("")), chunk("02"))},
		{"long bitlist filled", bitlist(repeat("T", 512), 512), repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), chunk("0002"))},
		{"odd bitvector filled", bitvec(repeat("T", 513)), repeat("ff", 64) + "01", h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk("")))},
		{"odd bitlist filled", bitlist(repeat("T", 513), 513), repeat("ff", 64) + "03", h(h(h(repeat("ff", 32), repeat("ff", 32)), h(chunk("01"), chunk(""))), chunk("0102"))},
		{"uint8 00", Uint8View(0x00), "00", chunk("00")},
		{"uint8 01", Uint8View(0x01), "01", chunk("01")},
		{"uint8 ab", Uint8View(0xab), "ab", chunk("ab")},

		{"byte 00", ByteView(0x00), "00", chunk("00")},
		{"byte 01", ByteView(0x01), "01", chunk("01")},
		{"byte ab", ByteView(0xab), "ab", chunk("ab")},

		{"uint16 0000", Uint16View(0x0000), "0000", chunk("0000")},
		{"uint16 abcd", Uint16View(0xabcd), "cdab", chunk("cdab")},
		{"uint32 00000000", Uint32View(0x00000000), "00000000", chunk("00000000")},
		{"uint32 01234567", Uint32View(0x01234567), "67452301", chunk("67452301")},
		{"small {4567, 0123}", viewMust(SmallTestStructType.FromFields(Uint16View(0x4567), Uint16View(0x0123))), "67452301", h(chunk("6745"), chunk("2301"))},
		{"small [4567, 0123]::2", viewMust(BasicVectorType(Uint16Type, 2).FromElements(Uint16View(0x4567), Uint16View(0x0123))), "67452301", chunk("67452301")},
		{"uint32 01234567", Uint32View(0x01234567), "67452301", chunk("67452301")},
		{"uint64 0000000000000000", Uint64View(0), "0000000000000000", chunk("0000000000000000")},
		{"uint64 0123456789abcdef", Uint64View(0x0123456789abcdef), "efcdab8967452301", chunk("efcdab8967452301")},
		{"uint256 0000000000000000000000000000000000000000000000000000000000000000", Uint256View{}, "0000000000000000000000000000000000000000000000000000000000000000", "0000000000000000000000000000000000000000000000000000000000000000"},
		{"uint256 f1f0e1e0d1d0c1c0b1b0a1a09190818071706160515041403130212011100100", MustUint256("0xf1f0e1e0d1d0c1c0b1b0a1a09190818071706160515041403130212011100100"), "0001101120213031404150516061707180819091a0a1b0b1c0c1d0d1e0e1f0f1", "0001101120213031404150516061707180819091a0a1b0b1c0c1d0d1e0e1f0f1"},
		// TODO: bytelist type/view that is not backed by a tree, but makes the tree on demand, possible optimization.
		//	("bytes48", Vector[byte, 48], Vector[byte, 48](*range(48)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f",
		// h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f00000000000000000000000000000000")),
		//("raw bytes48", ByteVector[48], ByteVector[48](*range(48)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f",
		// h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f00000000000000000000000000000000")),
		//("small empty bytelist", List[byte, 10], List[byte, 10](), "", h(chunk(""), chunk("00"))),
		//("big empty bytelist", List[byte, 2048], List[byte, 2048](), "", h(zero_hashes[6], chunk("00"))),
		//("raw small empty bytelist", ByteList[10], ByteList[10](), "", h(chunk(""), chunk("00"))),
		//("raw big empty bytelist", ByteList[2048], ByteList[2048](), "", h(zero_hashes[6], chunk("00"))),
		//("bytelist 7", List[byte, 7], List[byte, 7](*range(7)), "00010203040506",
		// h(chunk("00010203040506"), chunk("07"))),
		//("raw bytelist 7", ByteList[7], ByteList[7](*range(7)), "00010203040506",
		// h(chunk("00010203040506"), chunk("07"))),
		//("bytelist 50", List[byte, 50], List[byte, 50](*range(50)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031",
		// h(h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f30310000000000000000000000000000"), chunk("32"))),
		//("raw bytelist 50", ByteList[50], ByteList[50](*range(50)), "000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f202122232425262728292a2b2c2d2e2f3031",
		// h(h("000102030405060708090a0b0c0d0e0f101112131415161718191a1b1c1d1e1f", "202122232425262728292a2b2c2d2e2f30310000000000000000000000000000"), chunk("32"))),
		//("bytelist 6/256", List[byte, 256], List[byte, 256](*range(6)), "000102030405",
		// h(h(h(h(chunk("000102030405"), zero_hashes[0]), zero_hashes[1]), zero_hashes[2]), chunk("06"))),
		//("raw bytelist 6/256", ByteList[256], List[byte, 256](*range(6)), "000102030405",
		// h(h(h(h(chunk("000102030405"), zero_hashes[0]), zero_hashes[1]), zero_hashes[2]), chunk("06"))),
		//("sig", Vector[byte, 96], Vector[byte, 96](*sig_test_data),
		// "0100000000000000000000000000000000000000000000000000000000000000"
		// "0200000000000000000000000000000000000000000000000000000000000000"
		// "03000000000000000000000000000000000000000000000000000000000000ff",
		// h(h(chunk("01"), chunk("02")),
		//   h("03000000000000000000000000000000000000000000000000000000000000ff", chunk("")))),
		//{"sig", [96]byte{0: 1, 32: 2, 64: 3, 95: 0xff},
		//"01" + repeat("00", 31) + "02" + repeat("00", 31) + "03" + repeat("00", 30) + "ff",
		//h(h(chunk("01"), chunk("02")), h("03"+repeat("00", 30)+"ff", chunk("")))},
		//("raw sig", ByteVector[96], ByteVector[96](*sig_test_data),
		// "0100000000000000000000000000000000000000000000000000000000000000"
		// "0200000000000000000000000000000000000000000000000000000000000000"
		// "03000000000000000000000000000000000000000000000000000000000000ff",
		// h(h(chunk("01"), chunk("02")),
		//   h("03000000000000000000000000000000000000000000000000000000000000ff", chunk("")))),
		//("3 sigs", Vector[ByteVector[96], 3], Vector[ByteVector[96], 3](
		//    [1] + [0 for i in range(95)],
		//    [2] + [0 for i in range(95)],
		//    [3] + [0 for i in range(95)]
		//),
		// "01" + ("00" * 95) + "02" + ("00" * 95) + "03" + ("00" * 95),
		// h(h(h(h(chunk("01"), chunk("")), zero_hashes[1]), h(h(chunk("02"), chunk("")), zero_hashes[1])),
		//   h(h(h(chunk("03"), chunk("")), zero_hashes[1]), chunk("")))),
		{"singleFieldTestStruct", viewMust(SingleFieldTestStructType.FromFields(Uint16View(0xab))), "ab", chunk("ab")},

		{"uint16 list", viewMust(BasicListType(Uint16Type, 32).FromElements(Uint16View(0xaabb), Uint16View(0xc0ad), Uint16View(0xeeff))), "bbaaadc0ffee",
			h(h(chunk("bbaaadc0ffee"), chunk("")), chunk("03000000")), // max length: 32 * 2 = 64 bytes = 2 chunks
		},
		{"uint32 list", viewMust(BasicListType(Uint32Type, 128).FromElements(Uint32View(0xaabb), Uint32View(0xc0ad), Uint32View(0xeeff))), "bbaa0000adc00000ffee0000",
			// max length: 128 * 4 = 512 bytes = 16 chunks
			h(merge(chunk("bbaa0000adc00000ffee0000"), zeroHashes[0:4]), chunk("03000000")),
		},
		{"bytes32 list", viewMust(ComplexListType(RootType, 64).FromElements(&RootView{0xbb, 0xaa}, &RootView{0xad, 0xc0}, &RootView{0xff, 0xee})),
			"bbaa000000000000000000000000000000000000000000000000000000000000" +
				"adc0000000000000000000000000000000000000000000000000000000000000" +
				"ffee000000000000000000000000000000000000000000000000000000000000",
			h(merge(h(h(chunk("bbaa"), chunk("adc0")), h(chunk("ffee"), chunk(""))), zeroHashes[2:6]), chunk("03000000")),
		},
		{"uint256 list", viewMust(ComplexListType(Uint256Type, 64).FromElements(&Uint256View{0: 0xaabb}, &Uint256View{0: 0xc0ad}, &Uint256View{0: 0xeeff})),
			"bbaa000000000000000000000000000000000000000000000000000000000000" +
				"adc0000000000000000000000000000000000000000000000000000000000000" +
				"ffee000000000000000000000000000000000000000000000000000000000000",
			h(merge(h(h(chunk("bbaa"), chunk("adc0")), h(chunk("ffee"), chunk(""))), zeroHashes[2:6]), chunk("03000000")),
		},
		{"bytes32 list long", viewMust(ComplexListType(RootType, 128).FromElements(
			&RootView{1}, &RootView{2}, &RootView{3}, &RootView{4}, &RootView{5}, &RootView{6}, &RootView{7}, &RootView{8}, &RootView{9}, &RootView{10},
			&RootView{11}, &RootView{12}, &RootView{13}, &RootView{14}, &RootView{15}, &RootView{16}, &RootView{17}, &RootView{18}, &RootView{19},
		)),
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
		},
		{"fixedTestStruct", viewMust(FixedTestStructType.FromFields(Uint8View(0xab), Uint64View(0xaabbccdd00112233), Uint32View(0x12345678))), "ab33221100ddccbbaa78563412",
			h(h(chunk("ab"), chunk("33221100ddccbbaa")), h(chunk("78563412"), chunk("")))},
		{"VarTestStruct empty", viewMust(VarTestStructType.FromFields(Uint16View(0xabcd), BasicListType(Uint16Type, 1024).New(), Uint8View(0xff))), "cdab07000000ff",
			// log2(1024*2/32)= 6 deep
			h(h(chunk("cdab"), h(zeroHashes[6], chunk("00000000"))), h(chunk("ff"), chunk("")))},
		{"VarTestStruct some", viewMust(
			VarTestStructType.FromFields(
				Uint16View(0xabcd),
				viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(1), Uint16View(2), Uint16View(3))),
				Uint8View(0xff))),
			"cdab07000000ff010002000300",
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
		},
		{"empty list", ListAType.New(), "", h(zeroHashes[2], chunk("00000000"))},
		{"empty var element list", ListBType.New(), "", h(zeroHashes[3], chunk("00000000"))},
		{"var element list", viewMust(ComplexListType(VarTestStructType, 8).FromElements(
			viewMust(VarTestStructType.FromFields(Uint16View(0xdead), viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(1), Uint16View(2), Uint16View(3))), Uint8View(0x11))),
			viewMust(VarTestStructType.FromFields(Uint16View(0xbeef), viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(4), Uint16View(5), Uint16View(6))), Uint8View(0x22))),
		)),
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
			), chunk("02000000"))},
		{"empty list fields", ListStructType.New(), "08000000" + "08000000",
			h(h(zeroHashes[2], chunk("")), h(zeroHashes[3], chunk("")))},
		{"empty last field", viewMust(ListStructType.FromFields(viewMust(ListAType.FromElements(
			viewMust(SmallTestStructType.FromFields(Uint16View(0xaa11), Uint16View(0xbb22))),
			viewMust(SmallTestStructType.FromFields(Uint16View(0xcc33), Uint16View(0xdd44))),
			viewMust(SmallTestStructType.FromFields(Uint16View(0x1234), Uint16View(0x4567))),
		)), ListBType.New())), "08000000" + "14000000" + ("11aa22bb" + "33cc44dd" + "34126745"),
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
			)},
		{"complexTestStruct",
			viewMust(ComplexTestStructType.FromFields(
				Uint16View(0xaabb),
				viewMust(BasicListType(Uint16Type, 128).FromElements(Uint16View(0x1122), Uint16View(0x3344))),
				Uint8View(0xff),
				viewMust(BasicListType(ByteType, 256).FromElements(ByteView('f'), ByteView('o'), ByteView('o'), ByteView('b'), ByteView('a'), ByteView('r'))),
				viewMust(
					VarTestStructType.FromFields(
						Uint16View(0xabcd),
						viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(1), Uint16View(2), Uint16View(3))),
						Uint8View(0xff))),
				viewMust(ComplexVectorType(FixedTestStructType, 4).FromElements(
					viewMust(FixedTestStructType.FromFields(Uint8View(0xcc), Uint64View(0x4242424242424242), Uint32View(0x13371337))),
					viewMust(FixedTestStructType.FromFields(Uint8View(0xdd), Uint64View(0x3333333333333333), Uint32View(0xabcdabcd))),
					viewMust(FixedTestStructType.FromFields(Uint8View(0xee), Uint64View(0x4444444444444444), Uint32View(0x00112233))),
					viewMust(FixedTestStructType.FromFields(Uint8View(0xff), Uint64View(0x5555555555555555), Uint32View(0x44556677))),
				)),
				viewMust(ComplexVectorType(VarTestStructType, 2).FromElements(
					viewMust(VarTestStructType.FromFields(
						Uint16View(0xdead),
						viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(1), Uint16View(2), Uint16View(3))),
						Uint8View(0x11))),
					viewMust(VarTestStructType.FromFields(
						Uint16View(0xbeef),
						viewMust(BasicListType(Uint16Type, 1024).FromElements(Uint16View(4), Uint16View(5), Uint16View(6))),
						Uint8View(0x22))),
				)),
			)),
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
			)},
	}
}

func TestSerializeView(t *testing.T) {
	var buf bytes.Buffer

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			if err := tt.value.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
				t.Fatalf("encoding failed, err: %v", err)
			}
			data := buf.Bytes()
			if res := fmt.Sprintf("%x", data); res != tt.hex {
				t.Fatalf("encoded different data:\n     got %s\nexpected %s", res, tt.hex)
			}
		})
	}
}

func TestDeserializeSerialize(t *testing.T) {
	for _, tt := range testCases {
		v := reflect.New(reflect.TypeOf(tt.value))
		if _, ok := v.Interface().(codec.Deserializable); ok {
			t.Run(tt.name, func(t *testing.T) {
				d := v.Interface().(codec.Deserializable)
				data, err := hex.DecodeString(tt.hex)
				if err != nil {
					t.Fatal(err)
				}
				r := bytes.NewReader(data)
				// For dynamic types, we need to pass the length of the message to the decoder.
				bytesLen := uint64(len(tt.hex)) / 2
				if err := d.Deserialize(codec.NewDecodingReader(r, bytesLen)); err != nil {
					t.Fatalf("decoding failed, err: %v", err)
				}
				s, ok := d.(codec.Serializable)
				if !ok {
					t.Fatal("implemented Deserializable, but not Serializable")
				}
				var buf bytes.Buffer
				if err := s.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
					t.Fatal(err)
				}
				got := hex.EncodeToString(buf.Bytes())
				if got != tt.hex {
					t.Fatalf("got %s, expected %s", got, tt.hex)
				}
			})
		}
	}
}

func TestDecodeEncode(t *testing.T) {
	for _, tt := range testCases {
		v := reflect.New(reflect.TypeOf(tt.value))
		if _, ok := v.Interface().(codec.Decodable); ok {
			t.Run(tt.name, func(t *testing.T) {
				d := v.Interface().(codec.Decodable)
				data, err := hex.DecodeString(tt.hex)
				if err != nil {
					t.Fatal(err)
				}
				if err := d.Decode(data); err != nil {
					t.Fatalf("decoding failed, err: %v", err)
				}
				s, ok := d.(codec.Encodable)
				if !ok {
					t.Fatal("implemented Decodable, but not Encodable")
				}
				out, err := s.Encode()
				if err != nil {
					t.Fatal(err)
				}
				got := hex.EncodeToString(out)
				if got != tt.hex {
					t.Fatalf("got %s, expected %s", got, tt.hex)
				}
			})
		}
	}
}

func TestValueByteLength(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			expectedSize := uint64(len(tt.hex)) / 2
			if size, err := tt.value.ValueByteLength(); err != nil {
				t.Error(err)
			} else if expectedSize != size {
				t.Errorf("encoded output does not match expected size: %d but got: %d", expectedSize, size)
			}
		})
	}
}

func TestDeserializeView(t *testing.T) {
	hFn := tree.GetHashFn()
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sszTyp := tt.value.Type()
			data, err := hex.DecodeString(tt.hex)
			if err != nil {
				t.Fatal(err)
			}
			r := bytes.NewReader(data)
			// For dynamic types, we need to pass the length of the message to the decoder.
			bytesLen := uint64(len(tt.hex)) / 2

			dest, err := sszTyp.Deserialize(codec.NewDecodingReader(r, bytesLen))
			if err != nil {
				t.Fatal(err)
			}
			root := dest.HashTreeRoot(hFn)
			hexRoot := hex.EncodeToString(root[:])
			if hexRoot != tt.root {
				t.Errorf("Hash tree root of deserialized object does not match expected root. Got: %s, expected: %s.", hexRoot, tt.root)
			}
			var buf bytes.Buffer
			if err := dest.Serialize(codec.NewEncodingWriter(&buf)); err != nil {
				t.Errorf("failed to serialize object that was just deserialized: %v", err)
			}
			out := hex.EncodeToString(buf.Bytes())
			if out != tt.hex {
				t.Errorf("original bytes do not match bytes after deserialize/serialize round-trip:\nout: %s\nin:  %s", out, tt.hex)
			}
		})
	}
}

func TestHashTreeRoot(t *testing.T) {
	var buf bytes.Buffer

	hFn := tree.GetHashFn()

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			buf.Reset()
			root := tt.value.HashTreeRoot(hFn)
			res := hex.EncodeToString(root[:])
			if res != tt.root {
				t.Errorf("Expected root %s but got %s", tt.root, res)
			}
			root2 := tt.value.HashTreeRoot(hFn)
			if root2 != root {
				t.Errorf("Hash-tree-root cache is broken, got different root for same value after another HTR")
			}
		})
	}
}

func TestTypeBounds(t *testing.T) {
	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			sszTyp := tt.value.Type()
			bytesLen := uint64(len(tt.hex)) / 2

			min := sszTyp.MinByteLength()
			max := sszTyp.MaxByteLength()
			size := sszTyp.TypeByteLength()
			if sszTyp.IsFixedByteLength() {
				if size != bytesLen {
					t.Errorf("Fixed length object is different byte size than expected: got %d, expected %d", size, bytesLen)
				}
				if min != max || min != size {
					t.Errorf("expected min, max and size to be equal. Got: %d %d %d", min, max, size)
				}
			} else {
				if size != 0 {
					t.Errorf("variable length object has non-0 static size: %d", size)
				}
				if bytesLen < min || bytesLen > max {
					t.Errorf("bounds of variable length object are wrong: min: %d max: %d, expected size: %d", min, max, bytesLen)
				}
			}
		})
	}
}
