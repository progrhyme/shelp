package cli

import (
	"errors"
	"io"

	"github.com/progrhyme/claft/internal/config"
	"github.com/progrhyme/claft/internal/git"
)

var (
	ErrUsage         = errors.New("Usage is shown")
	ErrParseFailed   = errors.New("Cannot parse flags")
	ErrArgument      = errors.New("Invalid argument")
	ErrCommandFailed = errors.New("Command execution failed")
)

type Cli struct {
	version   string
	config    config.Config
	git       git.Git
	outWriter io.Writer
	errWriter io.Writer
}

type commonCmd struct {
	config  config.Config
	out     io.Writer
	err     io.Writer
	command string
}

type commonFlags struct {
	help *bool
}

func NewCli(ver string, cfg config.Config, g git.Git, out, err io.Writer) Cli {
	return Cli{version: ver, config: cfg, git: g, outWriter: out, errWriter: err}
}

func (c *Cli) ParseAndExec(args []string) error {
	prog := args[0]

	common := commonCmd{config: c.config, out: c.outWriter, err: c.errWriter, command: prog}
	root := newRootCmd(common, c.version)
	install := newInstallCmd(common, c.git)
	remove := newRemoveCmd(common)

	if len(args) == 1 {
		root.flags.Usage()
		return ErrUsage
	}

	switch args[1] {
	case "install":
		return install.parseAndExec(args[2:])
	case "remove":
		return remove.parseAndExec(args[2:])
	default:
		return root.parseAndExec(args[1:])
	}
}
