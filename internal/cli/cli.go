package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

var (
	ErrUsage           = errors.New("Usage is shown")
	ErrParseFailed     = errors.New("Cannot parse flags")
	ErrArgument        = errors.New("Invalid argument")
	ErrCommandFailed   = errors.New("Command execution failed")
	ErrOperationFailed = errors.New("Operation failed")
	ErrCanceled        = errors.New("Operation is canceled")
	ErrWarning         = errors.New("Warning")
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
	flags   pflag.FlagSet
	out     io.Writer
	err     io.Writer
	command string
}

func NewCli(ver string, cfg config.Config, g git.Git, out, err io.Writer) Cli {
	return Cli{version: ver, config: cfg, git: g, outWriter: out, errWriter: err}
}

func (c *Cli) ParseAndExec(args []string) error {
	prog := filepath.Base(args[0])

	common := commonCmd{config: c.config, out: c.outWriter, err: c.errWriter, command: prog}
	root := newRootCmd(common, c.version)
	initializer := newInitCmd(common)
	installer := newInstallCmd(common, c.git)
	lister := newListCmd(common)
	remover := newRemoveCmd(common)
	upgrader := newUpgradeCmd(common, c.git)
	destroyer := newDestroyCmd(common)

	if len(args) == 1 {
		root.flags.Usage()
		return ErrUsage
	}

	switch args[1] {
	case "init":
		return initializer.parseAndExec(args[2:])
	case "install", "add":
		return installer.parseAndExec(args[1:])
	case "list":
		return lister.parseAndExec(args[2:])
	case "remove", "uninstall":
		return remover.parseAndExec(args[1:])
	case "upgrade":
		return upgrader.parseAndExec(args[2:])
	case "destroy":
		return destroyer.parseAndExec(args[2:])
	default:
		return root.parseAndExec(args[1:])
	}
}

type flagger interface {
	helpFlg() *bool
}

type commonFlags struct {
	help *bool
}

func (flag *commonFlags) helpFlg() *bool {
	return flag.help
}

func parseStartHelp(
	flags *pflag.FlagSet, option flagger, output io.Writer, args []string, requireArg bool,
) (bool, error) {
	if requireArg && len(args) == 0 {
		flags.Usage()
		return true, ErrUsage
	}

	err := flags.Parse(args)
	if err != nil {
		fmt.Fprintf(output, "Error! %s\n", err)
		flags.Usage()
		return true, ErrParseFailed
	}

	if *option.helpFlg() {
		flags.Usage()
		return true, nil
	}

	if requireArg && flags.NArg() == 0 {
		flags.Usage()
		return true, ErrUsage
	}

	return false, nil
}
