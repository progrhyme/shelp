package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
)

type upgradeCmd struct {
	gitCmd
}

func newUpgradeCmd(common commonCmd, git git.Git) upgradeCmd {
	cmd := &upgradeCmd{}
	cmd.commonCmd = common
	cmd.git = git
	setupCmdFlags(cmd, "upgrade", cmd.usage)
	return *cmd
}

func (cmd *upgradeCmd) usage() {
	const help = `Summary:
  Upgrade installed packages.

Syntax:
  # Upgrade all installed packages
  {{.Prog}} {{.Cmd}}

  # Upgrade a single package
  {{.Prog}} {{.Cmd}} <package>

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "upgrade"})

	cmd.flags.PrintDefaults()
}

func (cmd *upgradeCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
	if done || err != nil {
		return err
	}

	if cmd.flags.NArg() == 0 {
		return cmd.upgradeAll()
	}

	return cmd.upgradeOne(cmd.flags.Arg(0))
}

func (cmd *upgradeCmd) upgradeOne(pkg string) error {
	path := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is not installed\n", pkg)
		return ErrArgument
	}

	if err := os.Chdir(path); err != nil {
		fmt.Fprintf(cmd.err, "Error! Directory change failed. Path = %s\n", path)
		return ErrOperationFailed
	}

	err := cmd.git.Pull(*cmd.option.verbose)
	if err != nil {
		return ErrCommandFailed
	}
	return nil
}

func (cmd *upgradeCmd) upgradeAll() error {
	pkgs, err := installedPackages(cmd, true)
	if err != nil {
		return err
	}

	upgraded := 0
	for _, pkg := range pkgs {
		old, err := hasPackageUpdate(cmd, pkg.Name())
		switch err {
		case nil:
			// do nothing
		case ErrOperationFailed:
			return err
		default:
			return ErrCommandFailed
		}

		if !old {
			continue
		}

		fmt.Fprintf(cmd.out, "Upgrading \"%s\" ...\n", pkg.Name())
		err = cmd.upgradeOne(pkg.Name())
		if err != nil {
			return err
		}
		upgraded++
	}

	if upgraded > 0 {
		fmt.Fprintf(cmd.out, "%d packages upgraded\n", upgraded)
	} else {
		fmt.Fprintln(cmd.out, "All packages are up-to-date")
	}
	return nil
}
