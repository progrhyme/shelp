package cli

import (
	"errors"
	"io"

	"github.com/progrhyme/claft/internal/config"
	"github.com/progrhyme/claft/internal/git"
)

type Cli struct {
	version string
	config  config.Config
	git     git.Git
	output  io.Writer
}

func NewCli(ver string, cfg config.Config, g git.Git, out io.Writer) Cli {
	return Cli{version: ver, config: cfg, git: g, output: out}
}

var (
	ErrUsage         = errors.New("Usage is shown")
	ErrParseFailed   = errors.New("Cannot parse flags")
	ErrArgument      = errors.New("Invalid argument")
	ErrCommandFailed = errors.New("Command execution failed")
)

func (c *Cli) ParseAndExec(args []string) error {
	prog := args[0]

	root := newRootCmd(c.output, prog, c.version)
	install := newInstallCmd(c.output, c.config, c.git, prog)

	if len(args) == 1 {
		root.flags.Usage()
		return ErrUsage
	}

	switch args[1] {
	case "install":
		return install.parseAndExec(args[2:])

	default:
		return root.parseAndExec(args[1:])
	}
}
