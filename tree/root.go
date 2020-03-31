package tree

type Root [32]byte

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
