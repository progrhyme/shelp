package cli

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

var (
	ErrConfig          = errors.New("Configuration error")
	ErrUsage           = errors.New("Usage is shown")
	ErrParseFailed     = errors.New("Cannot parse flags")
	ErrArgument        = errors.New("Invalid argument")
	ErrCommandFailed   = errors.New("Command execution failed")
	ErrOperationFailed = errors.New("Operation failed")
	ErrNoPackage       = errors.New("No package is installed")
	ErrCanceled        = errors.New("Operation is canceled")
	ErrWarning         = errors.New("Warning")
)

type Cli struct {
	version   string
	config    *config.Config
	git       git.Git
	outWriter io.Writer
	errWriter io.Writer
}

func NewCli(ver string, cfg *config.Config, g git.Git, out, err io.Writer) Cli {
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
	case "outdated":
		lister := newOutdatedCmd(common, c.git)
		return lister.parseAndExec(args[2:])
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

func setupCmdFlags(cmd interface{}, name string, usage func()) {
	flags := pflag.NewFlagSet(name, pflag.ContinueOnError)
	cmd.(commander).setFlags(flags)
	flags = cmd.(commander).flagset()
	flags.SetOutput(cmd.(commander).errs())
	if usage != nil {
		flags.Usage = usage
	}

	switch v := cmd.(type) {
	case verboseCommander:
		option := cmd.(verboseCommander).verboseOpts()
		option.setConfig(flags.StringP("config", "c", "", "# Configuration file"))
		option.setHelp(flags.BoolP("help", "h", false, "# Show help"))
		option.setVerbose(flags.BoolP("verbose", "v", false, "# Verbose output"))

	case helpCommander:
		option := cmd.(helpCommander).getOpts()
		option.setConfig(flags.StringP("config", "c", "", "# Configuration file"))
		option.setHelp(flags.BoolP("help", "h", false, "# Show help"))

	default:
		panic(fmt.Sprintf("Unexpected type! cmd: %v, type: %v", cmd, v))
	}
}

// Start parsing command-line arguments
// Then, load configuration file if it exists
func parseStart(cmd helpCommander, args []string, requireArg bool) (done bool, e error) {
	if requireArg && len(args) == 0 {
		cmd.flagset().Usage()
		return true, ErrUsage
	}

	if err := cmd.flagset().Parse(args); err != nil {
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

	// Load config file
	if err := cmd.getConf().LoadConfig(*cmd.getOpts().confFile()); err != nil {
		return true, ErrConfig
	}
	if cmd.getConf().IsLoaded() {
		fmt.Fprintf(cmd.errs(), "Use config: %s\n", cmd.getConf().File())
	}

	return false, nil
}

func installedPackages(cmd commander, noPkgErr bool) ([]os.FileInfo, error) {
	var pkgs []os.FileInfo
	nopkg := func() ([]os.FileInfo, error) {
		fmt.Fprintln(cmd.errs(), "No package is installed")
		if noPkgErr {
			return pkgs, ErrNoPackage
		} else {
			return pkgs, nil
		}
	}
	if _, err := os.Stat(cmd.getConf().PackagePath()); os.IsNotExist(err) {
		return nopkg()
	}

	pkgs, err := ioutil.ReadDir(cmd.getConf().PackagePath())
	if err != nil {
		fmt.Fprintf(cmd.errs(), "Error! %s\n", err)
		return pkgs, ErrOperationFailed
	}

	if len(pkgs) == 0 {
		return nopkg()
	}

	return pkgs, nil
}
