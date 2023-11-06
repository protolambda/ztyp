package tree

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"testing"
)

func TestMerkleProof(t *testing.T) {
	leaf := func(i uint64) Root {
		out := Root{}
		binary.BigEndian.PutUint64(out[24:], i+1)
		return out
	}
	proofs := func(i uint64, gIndex Gindex) []Root {
		return nil
	}
	tests := []struct {
		count    uint64
		limit    uint64
		index    uint64
		expected []string
	}{
		{8, 8, 0, []string{
			"f21a13e8c312a582f7a002f90b7af23bf104d4bfe70be85fe16229db8174c8f9",
			"373fade5331b44f7c3bbde6d110d5f62a9c658029567331fb29d2df285b8456a",
			"da0cd6b20f3f0d2e7c040ea5d411fed01cbe85af2bdc3069d7d191073b4bd2a7",
			"0000000000000000000000000000000000000000000000000000000000000002",
		}},

		{8, 9, 0, []string{
			"933df3778817647f2cedb0b2a43e15b2a27b138f7050132fa3d6f5ef5141c55a",
			"c78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c",
			"373fade5331b44f7c3bbde6d110d5f62a9c658029567331fb29d2df285b8456a",
			"da0cd6b20f3f0d2e7c040ea5d411fed01cbe85af2bdc3069d7d191073b4bd2a7",
			"0000000000000000000000000000000000000000000000000000000000000002",
		}},
		{9, 9, 8, []string{
			"b00c4fa7e5c2f468627398409af7bc6fb364c8bbcf629d58ae74385e08aee694",
			"f21a13e8c312a582f7a002f90b7af23bf104d4bfe70be85fe16229db8174c8f9",
			"db56114e00fdd4c1f85c892bf35ac9a89289aaecb1ebd0a96cde606a748b5d71",
			"f5a5fd42d16a20302798ef6ed309979b43003d2320d9f0e8ea9831a92759fb4b",
			"0000000000000000000000000000000000000000000000000000000000000000",
		}},
		{6, 1024, 0, []string{
			"f6b1035624bd2bb134a52924e41d66c8e31177458809de4ac8a2b0ac2ae1cd41",
			"506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1",
			"26846476fd5fc54a5d43385167c95144f2643f533cc85bb9d16b782f8d7db193",
			"87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c",
			"d88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1",
			"9efde052aa15429fae05bad4d0b1d7c64da64d03d7a1854a588c2cb8430c0d30",
			"536d98837f2dd165a55d5eeae91485954472d56f246df256bf3cae19352a123c",
			"c78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c",
			"9f9315983e72b3712b8453057fffea967876477459f4f288a79ac135f7e7ea7f",
			"da0cd6b20f3f0d2e7c040ea5d411fed01cbe85af2bdc3069d7d191073b4bd2a7",
			"0000000000000000000000000000000000000000000000000000000000000002",
		}},
		{6, 1024, 1, []string{
			"f6b1035624bd2bb134a52924e41d66c8e31177458809de4ac8a2b0ac2ae1cd41",
			"506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1",
			"26846476fd5fc54a5d43385167c95144f2643f533cc85bb9d16b782f8d7db193",
			"87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c",
			"d88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1",
			"9efde052aa15429fae05bad4d0b1d7c64da64d03d7a1854a588c2cb8430c0d30",
			"536d98837f2dd165a55d5eeae91485954472d56f246df256bf3cae19352a123c",
			"c78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c",
			"9f9315983e72b3712b8453057fffea967876477459f4f288a79ac135f7e7ea7f",
			"da0cd6b20f3f0d2e7c040ea5d411fed01cbe85af2bdc3069d7d191073b4bd2a7",
			"0000000000000000000000000000000000000000000000000000000000000001",
		}},
		{6, 1023, 1, []string{
			"f6b1035624bd2bb134a52924e41d66c8e31177458809de4ac8a2b0ac2ae1cd41",
			"506d86582d252405b840018792cad2bf1259f1ef5aa5f887e13cb2f0094f51e1",
			"26846476fd5fc54a5d43385167c95144f2643f533cc85bb9d16b782f8d7db193",
			"87eb0ddba57e35f6d286673802a4af5975e22506c7cf4c64bb6be5ee11527f2c",
			"d88ddfeed400a8755596b21942c1497e114c302e6118290f91e6772976041fa1",
			"9efde052aa15429fae05bad4d0b1d7c64da64d03d7a1854a588c2cb8430c0d30",
			"536d98837f2dd165a55d5eeae91485954472d56f246df256bf3cae19352a123c",
			"c78009fdf07fc56a11f122370658a353aaa542ed63e44c4bc15ff4cd105ab33c",
			"9f9315983e72b3712b8453057fffea967876477459f4f288a79ac135f7e7ea7f",
			"da0cd6b20f3f0d2e7c040ea5d411fed01cbe85af2bdc3069d7d191073b4bd2a7",
			"0000000000000000000000000000000000000000000000000000000000000001",
		}},
	}
	for _, test := range tests {
		t.Run("", func(t *testing.T) {
			gindex, err := ToGindex64(test.index, CoverDepth(test.limit))
			if err != nil {
				t.Fatal(err)
			}
			got := MerkleProof(Hash, test.count, test.limit, gindex, leaf, proofs)
			if len(got) != len(test.expected) {
				t.Fatalf("got %d items, expected %d", len(got), len(test.expected))
			}
			root := Merkleize(Hash, test.count, test.limit, leaf)
			if !bytes.Equal(root[:], got[0][:]) {
				t.Errorf("unexpected root, got %x, expected %x", got[len(got)-1][:], root[:])
			}
			for i, v := range got {
				expectedBytes, err := hex.DecodeString(test.expected[i])
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(v[:], expectedBytes) {
					t.Errorf("at index %d, got %x, expected %x", i, v[:], expectedBytes[:])
				}
			}
		})
	}
}
