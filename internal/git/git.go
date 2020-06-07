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

func (g *Git) Clone(src, dst string, verbose bool) error {
	args := []string{"clone", src, dst}
	if g.shallow {
		args = append(args, "--depth=1")
	}
	cmd := exec.Command(g.cmd, args...)
	if verbose {
		cmd.Stdout = g.out
		cmd.Stderr = g.err
	} else {
		cmd.Stdout = ioutil.Discard
		cmd.Stderr = ioutil.Discard
	}
	if verbose {
		fmt.Fprintln(cmd.Stdout, cmd.String())
	}

	return cmd.Run()
}
