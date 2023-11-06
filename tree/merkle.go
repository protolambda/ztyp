package tree

// Merkleize with log(N) space allocation
func Merkleize(hasher HashFn, count uint64, limit uint64, leaf func(i uint64) Root) (out Root) {
	if count > limit {
		// merkleizing list that is too large, over limit
		count = limit
	}
	if limit == 0 {
		return
	}
	if limit == 1 {
		if count == 1 {
			out = leaf(0)
		}
		return
	}
	depth := CoverDepth(count)
	limitDepth := CoverDepth(limit)
	tmp := make([]Root, limitDepth+1, limitDepth+1)

	j := uint8(0)
	hArr := Root{}

	merge := func(i uint64) {
		// merge back up from bottom to top, as far as we can
		for j = 0; ; j++ {
			// stop merging when we are in the left side of the next combi
			if i&(uint64(1)<<j) == 0 {
				// if we are at the count, we want to merge in zero-hashes for padding
				if i == count && j < depth {
					hArr = hasher(hArr, ZeroHashes[j])
				} else {
					break
				}
			} else {
				// keep merging up if we are the right side
				hArr = hasher(tmp[j], hArr)
			}
		}
		// store the merge result (may be no merge, i.e. bottom leaf node)
		tmp[j] = hArr
	}

	// merge in leaf by leaf.
	for i := uint64(0); i < count; i++ {
		hArr = leaf(i)
		merge(i)
	}

	// complement with 0 if empty, or if not the right power of 2
	if (uint64(1) << depth) != count {
		hArr = ZeroHashes[0]
		merge(count)
	}

	// the next power of two may be smaller than the ultimate virtual size,
	// complement with zero-hashes at each depth.
	for j := depth; j < limitDepth; j++ {
		tmp[j+1] = hasher(tmp[j], ZeroHashes[j])
	}

	return tmp[limitDepth]
}

// Compute the merkle proof of the given leaf index.
func MerkleProof(hasher HashFn, count uint64, limit uint64, gIndexFull Gindex, leaf func(i uint64) Root, proofs func(i uint64, gIndex Gindex) []Root) (out []Root) {
	if count > limit {
		// merkleizing list that is too large, over limit
		count = limit
	}
	if limit == 0 {
		return
	}
	depth := CoverDepth(count)
	limitDepth := CoverDepth(limit)
	if limit == 1 {
		out = make([]Root, 1)
		if count == 1 {
			out[0] = leaf(0)
		} else {
			out[0] = ZeroHashes[0]
		}
		return
	}

	out = make([]Root, limitDepth+1)

	gIndex, gIndexSubtree := gIndexFull.Split(uint32(limitDepth))

	type node struct {
		Root
		Gindex
	}
	newNode := func(r Root, g Gindex) *node {
		// Save the proof if necessary
		if g.IsProof(gIndex) {
			out[uint32(limitDepth)-g.Depth()] = r
		}
		return &node{r, g}
	}
	newLeafNode := func(i uint64) *node {
		leafGIndex, _ := ToGindex64(i, limitDepth)
		if i >= count {
			return newNode(ZeroHashes[0], leafGIndex)
		} else {
			return newNode(leaf(i), leafGIndex)
		}
	}
	nodeMerge := func(n1, n2 *node) *node {
		return newNode(hasher(n1.Root, n2.Root), n1.Parent())
	}

	tmp := make([]*node, limitDepth+1)
	j := uint8(0)
	hArr := &node{}

	merge := func(i uint64) {
		// merge back up from bottom to top, as far as we can
		for j = 0; ; j++ {
			// stop merging when we are in the left side of the next combi
			if i&(uint64(1)<<j) == 0 {
				// if we are at the count, we want to merge in zero-hashes for padding
				if i == count && j < depth {
					hArr = nodeMerge(hArr, newNode(ZeroHashes[j], hArr.Gindex.Sibling()))
				} else {
					break
				}
			} else {
				// keep merging up if we are the right side
				hArr = nodeMerge(tmp[j], hArr)
			}
		}
		// store the merge result (may be no merge, i.e. bottom leaf node)
		tmp[j] = hArr
	}

	// merge in leaf by leaf.
	for i := uint64(0); i < count; i++ {
		hArr = newLeafNode(i)
		merge(i)
	}

	// complement with 0 if empty, or if not the right power of 2
	if (uint64(1) << depth) != count {
		hArr = newLeafNode(count)
		merge(count)
	}

	// the next power of two may be smaller than the ultimate virtual size,
	// complement with zero-hashes at each depth.
	for j := depth; j < limitDepth; j++ {
		tmp[j+1] = nodeMerge(tmp[j], newNode(ZeroHashes[j], tmp[j].Gindex.Sibling()))
	}

	if !gIndexSubtree.IsRoot() {
		if childRoots := proofs(gIndex.BaseIndex(), gIndexSubtree); len(childRoots) > 0 {
			out = append(childRoots[:len(childRoots)-1], out[:len(out)-1]...)
		}
	}

	return
}

func VerifyProof(hasher HashFn, branch []Root, index Gindex, root Root, leaf Root) bool {
	if len(branch) != int(index.Depth()) {
		return false
	}
	for _, proof := range branch {
		if index.IsLeftLeaf() {
			leaf = hasher(leaf, proof)
		} else {
			leaf = hasher(proof, leaf)
		}
		index = index.Parent()
	}
	return leaf == root
}
