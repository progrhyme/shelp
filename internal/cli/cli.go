package cli

import (
	"errors"
	"fmt"
	"io"
	"path/filepath"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
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

func NewCli(ver string, cfg config.Config, g git.Git, out, err io.Writer) Cli {
	return Cli{version: ver, config: cfg, git: g, outWriter: out, errWriter: err}
}

func (c *Cli) ParseAndExec(args []string) error {
	prog := filepath.Base(args[0])

	common := commonCmd{config: c.config, out: c.outWriter, err: c.errWriter, command: prog}
	root := newRootCmd(common, c.version)

	if len(args) == 1 {
		root.flags.Usage()
		return ErrUsage
	}

	switch args[1] {
	case "init":
		initializer := newInitCmd(common)
		return initializer.parseAndExec(args[2:])
	case "install", "add":
		installer := newInstallCmd(common, c.git)
		return installer.parseAndExec(args[1:])
	case "list":
		lister := newListCmd(common)
		return lister.parseAndExec(args[2:])
	case "remove", "uninstall":
		remover := newRemoveCmd(common)
		return remover.parseAndExec(args[1:])
	case "upgrade":
		upgrader := newUpgradeCmd(common, c.git)
		return upgrader.parseAndExec(args[2:])
	case "link":
		linker := newLinkCmd(common)
		return linker.parseAndExec(args[2:])
	case "destroy":
		destroyer := newDestroyCmd(common)
		return destroyer.parseAndExec(args[2:])
	default:
		return root.parseAndExec(args[1:])
	}
}

func parseStartHelp(cmd helpCommander, args []string, requireArg bool) (bool, error) {
	if requireArg && len(args) == 0 {
		cmd.flagset().Usage()
		return true, ErrUsage
	}

	err := cmd.flagset().Parse(args)
	if err != nil {
		fmt.Fprintf(cmd.errs(), "Error! %s\n", err)
		cmd.flagset().Usage()
		return true, ErrParseFailed
	}

	if *cmd.getOpts().helpFlg() {
		cmd.flagset().Usage()
		return true, nil
	}

	if requireArg && cmd.flagset().NArg() == 0 {
		cmd.flagset().Usage()
		return true, ErrUsage
	}

	return false, nil
}
