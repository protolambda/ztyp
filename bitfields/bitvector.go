package bitfields

import "fmt"

// Helper function to implement Bitvector with.
// It checks if:
//  1. b has the same amount of bytes as necessary for n bits.
//  2. unused bits in b are 0
func BitvectorCheck(b []byte, n uint64) error {
	byteLen := uint64(len(b))
	if err := BitvectorCheckByteLen(byteLen, n); err != nil {
		return err
	}
	if byteLen == 0 {
		// empty bitvector
		return nil
	}
	last := b[byteLen-1]
	return BitvectorCheckLastByte(last, n)
}

func BitvectorCheckByteLen(byteLen uint64, bitLength uint64) error {
	if expected := (bitLength + 7) >> 3; byteLen != expected {
		return fmt.Errorf("bitvector of %d bytes has not expected length in bytes %d", byteLen, expected)
	}
	return nil
}

func BitvectorCheckLastByte(last byte, n uint64) error {
	if n == 0 {
		return fmt.Errorf("empty bitvector can not have a last byte")
	}
	if n&7 == 0 {
		// n is a multiple of 8, so last byte fits
		return nil
	}
	// check if it is big enough to hold any non-zero contents in the last byte.
	if expectedBitsLen := n & 7; (last >> expectedBitsLen) != 0 {
		return fmt.Errorf("bitvector last byte 0b%b has not expected %d bits in last byte", last, expectedBitsLen)
	}
	return nil
}
