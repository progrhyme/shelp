package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

// Git command executor
type Git struct {
	cmd string
	out io.Writer
	err io.Writer
}

// Option for some operations to pass from outside of this package
type Option struct {
	Branch  string
	Shallow bool
	Verbose bool
}

func NewGit(out, err io.Writer) Git {
	cmd := os.Getenv("GIT_COMMAND")
	if cmd == "" {
		cmd = "git"
	}
	return Git{cmd: cmd, out: out, err: err}
}

func (g *Git) Clone(src, dst string, opts Option) error {
	args := []string{"clone", src}
	if opts.Branch != "" {
		args = append(args, fmt.Sprintf("--branch=%s", opts.Branch))
	}
	if opts.Shallow {
		args = append(args, "--depth=1")
	}
	args = append(args, dst)
	return g.prepareCommand(args, opts.Verbose).Run()
}

// HasUpdate is supposed to be executed inside a working tree
func (g *Git) HasUpdate(verbose bool) (bool, error) {
	err := g.prepareCommand([]string{"fetch"}, verbose).Run()
	if err != nil {
		fmt.Fprintf(g.err, "Error! git fetch failed. Error = %v", err)
		return false, err
	}
	args := []string{"symbolic-ref", "--short", "--quiet", "HEAD"}
	if err = g.prepareCommand(args, false).Run(); err != nil {
		// Probably detached HEAD, no need to update
		return false, nil
	}

	args = []string{"rev-list", "--count", "HEAD...HEAD@{upstream}"}
	s, err := g.getCommandOutput(args, verbose, false)
	if err != nil {
		return false, err
	}

	var i int
	_, err = fmt.Fscanln(strings.NewReader(s.String()), &i)
	if err != nil {
		fmt.Fprintf(g.err, "Error! Cast failed: %v -> int. Error = %v", s, err)
		return false, err
	}

	if i > 0 {
		return true, nil
	} else {
		return false, nil
	}
}

func (g *Git) Pull(verbose bool) error {
	// NOTE: "shallow" option for Pull operation is incomplete.
	//  It happens to cause merge conflicts.
	//if g.shallow {
	//	cmd := g.prepareCommand([]string{"fetch", "--depth=1"}, verbose)
	//	if err := cmd.Run(); err != nil {
	//		return err
	//	}
	//}
	cmd := exec.Command(g.cmd, []string{"pull"}...)
	cmd.Stdout = g.out
	cmd.Stderr = g.err
	if verbose {
		fmt.Fprintf(g.err, "[CMD] %s\n", cmd.String())
	}

	return cmd.Run()
}

// Worktree is supposed to be executed inside a working tree
func (g *Git) Worktree(verbose bool) (Worktree, error) {
	wt := Worktree{}

	getCmdOut := func(args []string) string {
		s, err := g.getCommandOutput(args, verbose, true)
		if err == nil {
			return strings.TrimRight(s.String(), "\r\n")
		}
		return ""
	}
	wt.RemoteURL = getCmdOut([]string{"config", "--get", "remote.origin.url"})
	wt.Branch = getCmdOut([]string{"symbolic-ref", "--short", "--quiet", "HEAD"})
	wt.Tag = getCmdOut([]string{"tag", "--points-at", "HEAD"})
	defbranch := getCmdOut([]string{"symbolic-ref", "--short", "--quiet", "refs/remotes/origin/HEAD"})
	wt.defaultBranch = filepath.Base(defbranch)

	return wt, nil
}

func (g *Git) prepareCommand(args []string, verbose bool) *exec.Cmd {
	cmd := exec.Command(g.cmd, args...)
	if verbose {
		cmd.Stdout = g.out
		cmd.Stderr = g.err
		fmt.Fprintf(g.err, "[CMD] %s\n", cmd.String())
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}
	return cmd
}

func (g *Git) getCommandOutput(args []string, verbose, suppressError bool) (*strings.Builder, error) {
	s := &strings.Builder{}
	cmd := exec.Command(g.cmd, args...)
	cmd.Stdout = s
	if suppressError {
		cmd.Stderr = ioutil.Discard
	} else {
		cmd.Stderr = g.err
	}
	if verbose {
		fmt.Fprintf(g.err, "[CMD] %s\n", cmd.String())
	}
	if err := cmd.Run(); err != nil {
		if !suppressError {
			fmt.Fprintf(g.err, "Error! git command failed. Args = %v, Error = %v", args, err)
		}
		return s, err
	}
	return s, nil
}
