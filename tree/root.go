package tree

import (
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/protolambda/ztyp/codec"
)

type Root [32]byte

func (r Root) MarshalText() ([]byte, error) {
	return []byte("0x" + hex.EncodeToString(r[:])), nil
}

func (r Root) String() string {
	return "0x" + hex.EncodeToString(r[:])
}

func (r *Root) Deserialize(dr *codec.DecodingReader) error {
	if r == nil {
		return errors.New("nil root")
	}
	_, err := dr.Read(r[:])
	return err
}

func (r Root) HashTreeRoot(_ HashFn) Root {
	return r
}

func (r *Root) UnmarshalText(text []byte) error {
	if r == nil {
		return errors.New("cannot decode into nil Root")
	}
	if len(text) >= 2 && text[0] == '0' && (text[1] == 'x' || text[1] == 'X') {
		text = text[2:]
	}
	if len(text) != 64 {
		return fmt.Errorf("unexpected length string '%s'", string(text))
	}
	_, err := hex.Decode(r[:], text)
	return err
}

func (r *Root) Left() (Node, error) {
	return nil, NavigationError
}

func (r *Root) Right() (Node, error) {
	return nil, NavigationError
}

func (r *Root) IsLeaf() bool {
	return true
}

func (r *Root) RebindLeft(v Node) (Node, error) {
	return nil, NavigationError
}

func (r *Root) RebindRight(v Node) (Node, error) {
	return nil, NavigationError
}

func (r *Root) Getter(target Gindex) (Node, error) {
	if target.IsRoot() {
		return r, nil
	} else {
		return nil, NavigationError
	}
}

func (r *Root) Setter(target Gindex, expand bool) (Link, error) {
	if target.IsRoot() {
		return Identity, nil
	}
	if expand {
		child := ZeroNode(target.Depth())
		p := NewPairNode(child, child)
		return p.Setter(target, expand)
	} else {
		return nil, NavigationError
	}
}

func (r *Root) SummarizeInto(target Gindex, h HashFn) (SummaryLink, error) {
	if target.IsRoot() {
		return func() (Node, error) {
			return r, nil
		}, nil
	} else {
		return nil, NavigationError
	}
}

func (r *Root) MerkleRoot(h HashFn) Root {
	if r == nil {
		return Root{}
	}
	return *r
}
