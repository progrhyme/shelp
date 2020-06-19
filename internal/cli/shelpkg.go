package cli

// shelp package params
type shelpkg struct {
	name            string
	url             string
	ref             string
	isBranchDefault bool
	isCommitHash    bool
}

func (pkg *shelpkg) isEquivalent(tgt shelpkg) bool {
	if pkg.url != tgt.url {
		return false
	}
	if pkg.ref == tgt.ref {
		return true
	}
	if pkg.ref == "" && tgt.isBranchDefault {
		return true
	}
	return false
}
