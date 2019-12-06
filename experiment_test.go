package main

import (
	"github.com/protolambda/ztyp/tree"
	. "github.com/protolambda/ztyp/view"
	"testing"
)

var BLSPubkeyType = BasicVectorType(ByteType, 48)

var ValidatorType = &ContainerType{
	{"pubkey", BLSPubkeyType},
	{"withdrawal_credentials", RootType}, // Commitment to pubkey for withdrawals
	{"effective_balance", Uint64Type},    // Balance at stake
	{"slashed", BoolType},
	// Status epochs
	{"activation_eligibility_epoch", Uint64Type}, // When criteria for activation were met
	{"activation_epoch", Uint64Type},
	{"exit_epoch", Uint64Type},
	{"withdrawable_epoch", Uint64Type}, // When validator can withdraw funds
}

const VALIDATOR_REGISTRY_LIMIT uint64 = 1 << 40

var RegistryBalancesType = BasicListType(Uint64Type, VALIDATOR_REGISTRY_LIMIT)

var RegistryValidatorsType = ListType(ValidatorType, VALIDATOR_REGISTRY_LIMIT)

func BenchmarkRegInitHash(t *testing.B) {
	startCount := 100000
	regView := RegistryValidatorsType.New()
	for i := 0; i < startCount; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Error(err)
		}
	}
	for i := 0; i < t.N; i++ {
		r := regView.ViewRoot(tree.Hash)
		t.Logf("x; %x", r)
		ll, err := regView.Length()
		t.Logf("length: %d %v", ll, err)
	}
}

func BenchmarkRegHash(t *testing.B) {
	startCount := 100000
	regView := RegistryValidatorsType.New()
	r := regView.ViewRoot(tree.Hash)
	t.Logf("r; %x", r)
	for i := 0; i < startCount; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Error(err)
		}
	}
	ll, err := regView.Length()
	t.Logf("length: %d %v", ll, err)
	t.ResetTimer()
	for i := 0; i < t.N; i++ {
		if err := regView.Append(ValidatorType.New()); err != nil {
			t.Error(err)
		}
		r := regView.ViewRoot(tree.Hash)
		t.Logf("x; %x", r)
	}
}
