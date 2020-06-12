package git

import (
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
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
