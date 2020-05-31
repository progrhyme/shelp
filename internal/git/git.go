package git

import (
	"fmt"
	"io"
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

func (g *Git) Clone(src, dst string) error {
	args := []string{"clone", src, dst}
	if g.shallow {
		args = append(args, "--depth=1")
	}
	cmd := exec.Command(g.cmd, args...)
	cmd.Stdout = g.out
	cmd.Stderr = g.err
	// Debug. TODO: remove
	fmt.Fprintln(cmd.Stdout, cmd.String())

	return cmd.Run()
}
