package git

type Worktree struct {
	RemoteURL     string
	Branch        string
	Tag           string
	defaultBranch string
}

func (wt *Worktree) BranchOrTag() string {
	if wt.Branch != "" {
		return wt.Branch
	}
	return wt.Tag
}

func (wt *Worktree) IsBranchDefault() bool {
	return wt.Branch == wt.defaultBranch
}
