package tree

import (
	"crypto/sha256"
	"fmt"
)

type HashFn func(a Root, b Root) Root

func sha256Combi(a Root, b Root) Root {
	v := [64]byte{}
	copy(v[:32], a[:])
	copy(v[32:], b[:])
	return sha256.Sum256(v[:])
}

func sha256CombiRepeat() HashFn {
	hash := sha256.New()
	v := [64]byte{}
	hashFn := func(a Root, b Root) (out Root) {
		hash.Reset()
		copy(v[:32], a[:])
		copy(v[32:], b[:])
		hash.Write(v[:])
		copy(out[:], hash.Sum(nil))
		return
	}
	return hashFn
}

type NewHashFn func() HashFn

// Get a hash-function that re-uses the hashing working-variables. Defaults to SHA-256.
var GetHashFn NewHashFn = sha256CombiRepeat

var Hash HashFn = sha256Combi

var ZeroHashes []Root

// initialize the zero-hashes pre-computed data with the given hash-function.
func InitZeroHashes(h HashFn, zeroHashesLevels uint) {
	ZeroHashes = make([]Root, zeroHashesLevels+1)
	for i := uint(0); i < zeroHashesLevels; i++ {
		ZeroHashes[i+1] = h(ZeroHashes[i], ZeroHashes[i])
	}
}

func init() {
	InitZeroHashes(sha256Combi, 64)
}

func ZeroNode(depth uint32) Node {
	if depth >= uint32(len(ZeroHashes)) {
		panic(fmt.Errorf("depth %d reaches deeper than available %d precomputed zero-hashes provide", depth, len(ZeroHashes)))
	}
	return &ZeroHashes[depth]
}
