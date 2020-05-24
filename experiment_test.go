package main

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"testing"
)

var BLSPubkeyType = BasicVectorType(ByteType, 48)

var ValidatorType = ContainerType("Validator", []FieldDef{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", RootType}, // Commitment to pubkey for withdrawals
	{"effective_balance", Uint64Type},    // Balance at stake
	{"slashed", BoolType},
	// Status epochs
	{"activation_eligibility_epoch", Uint64Type}, // When criteria for activation were met
	{"activation_epoch", Uint64Type},
	{"exit_epoch", Uint64Type},
	{"withdrawable_epoch", Uint64Type}, // When validator can withdraw funds
})

const VALIDATOR_REGISTRY_LIMIT uint64 = 1 << 40

var RegistryBalancesType = BasicListType(Uint64Type, VALIDATOR_REGISTRY_LIMIT)

var RegistryValidatorsType = ComplexListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

func BenchmarkRegInitHash(t *testing.B) {
	startCount := 100000
	regView := RegistryValidatorsType.New()
	for i := 0; i < startCount; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Fatal(err)
		}
	}
	// after initial cost, hashing is cached, and it's essentially free.
	hFn := tree.GetHashFn()
	res := byte(0)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		r := regView.HashTreeRoot(hFn)
		res ^= r[0] // do something with the output, don't ignore it.
		//t.Logf("x; %x", r)
	}
	t.Logf("res %d", res)
	ll, err := regView.Length()
	t.Logf("length: %d %v, N: %d", ll, err, t.N)
}

func BenchmarkRegHash(t *testing.B) {
	startCount := 100000
	regView := RegistryValidatorsType.New()
	hFn := tree.GetHashFn()
	r := regView.HashTreeRoot(hFn)
	t.Logf("r; %x", r)
	for i := 0; i < startCount; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Fatal(err)
		}
	}
	ll, err := regView.Length()
	t.Logf("length: %d %v", ll, err)
	t.Logf("N: %d", t.N)
	t.ResetTimer()
	res := byte(0)
	for i := 0; i < t.N; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Error(err)
		}
		r := regView.HashTreeRoot(hFn)
		res ^= r[0] // do something with the output, don't ignore it.
		//t.Logf("x; %x", r)
	}
	t.Logf("res %d", res)
}
