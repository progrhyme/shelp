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
	t.Execute(cmd.errs, struct{ Prog, Cmd string }{cmd.name, "upgrade"})

	cmd.flags.PrintDefaults()
}

func (cmd *upgradeCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false, false)
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
		fmt.Fprintf(cmd.errs, "\"%s\" is not installed\n", pkg)
		return ErrArgument
	}

	pwd, err := chdir(cmd, path)
	if err != nil {
		return ErrOperationFailed
	}
	defer os.Chdir(pwd)

	hasUpdate, err := cmd.git.HasUpdate(*cmd.option.verbose)
	if err != nil {
		return ErrCommandFailed
	}

	if !hasUpdate {
		fmt.Fprintln(cmd.outs, "No need to upgrade")
		return nil
	}

	err = cmd.git.Pull(*cmd.option.verbose)
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

		fmt.Fprintf(cmd.outs, "Upgrading \"%s\" ...\n", pkg.Name())
		err = cmd.upgradeOne(pkg.Name())
		if err != nil {
			return err
		}
		upgraded++
	}

	if upgraded > 0 {
		fmt.Fprintf(cmd.outs, "%d packages upgraded\n", upgraded)
	} else {
		fmt.Fprintln(cmd.outs, "All packages are up-to-date")
	}
	return nil
}
