package conv

import (
	"errors"
	"strconv"
)

var DestNilErr = errors.New("destination is nil")
var EmptyInputErr = errors.New("input is empty")
var MissingQuoteErr = errors.New("input has quote open without close")

// Parse a uint64, with or without quotes, in any base, with common prefixes accepted to change base.
func Uint64Unmarshal(v *uint64, b []byte) error {
	if v == nil {
		return DestNilErr
	}
	if len(b) == 0 {
		return EmptyInputErr
	}
	if b[0] == '"' {
		if len(b) == 1 || b[len(b)-1] != '"' {
			return MissingQuoteErr
		}
		b = b[1 : len(b)-1]
	}
	n, err := strconv.ParseUint(string(b), 0, 64)
	if err != nil {
		return err
	}
	*v = n
	return nil
}

// Marshal a uint64, with quotes
func Uint64Marshal(v uint64) ([]byte, error) {
	var dest [18]byte
	dest[0] = '"'
	res := strconv.AppendUint(dest[0:1], v, 10)
	res = append(res, '"')
	return res, nil
}
