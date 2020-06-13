package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

type Git struct {
	cmd     string
	out     io.Writer
	err     io.Writer
	shallow bool
}

func NewGit(out, err io.Writer) Git {
	cmd := os.Getenv("GIT_COMMAND")
	if cmd == "" {
		cmd = "git"
	}
	return Git{cmd: cmd, out: out, err: err, shallow: true}
}

func (g *Git) Clone(src, dst, branch string, verbose bool) error {
	args := []string{"clone", src}
	if branch != "" {
		args = append(args, fmt.Sprintf("--branch=%s", branch))
	}
	if g.shallow {
		args = append(args, "--depth=1")
	}
	args = append(args, dst)
	return g.prepareCommand(args, verbose).Run()
}

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
	s := &strings.Builder{}
	cmd := exec.Command(g.cmd, args...)
	cmd.Stdout = s
	cmd.Stderr = g.err
	if verbose {
		fmt.Fprintf(g.err, "[CMD] %s\n", cmd.String())
	}
	err = cmd.Run()
	if err != nil {
		fmt.Fprintf(g.err, "Error! git rev-list failed. Error = %v", err)
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
