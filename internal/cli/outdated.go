package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
)

type outdatedCmd struct {
	gitCmd
}

func newOutdatedCmd(common commonCmd, git git.Git) outdatedCmd {
	cmd := &outdatedCmd{}
	cmd.commonCmd = common
	cmd.git = git
	setupCmdFlags(cmd, "outdated", cmd.usage)
	return *cmd
}

func (cmd *outdatedCmd) usage() {
	const help = `Summary:
  Show installed packages which can be updated.

Syntax:
  {{.Prog}} {{.Cmd}}

To update a package, run "{{.Prog}} upgrade <package>".

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "outdated"})

	cmd.flags.PrintDefaults()
}

func (cmd *outdatedCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(cmd, args, false)
	if done || err != nil {
		return err
	}

	pkgs, err := installedPackages(cmd, true)
	if err != nil {
		return err
	}

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

		if old {
			fmt.Fprintln(cmd.out, pkg.Name())
		} else {
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.err, "%s is up-to-date\n", pkg.Name())
			}
		}
	}

	return nil
}

func hasPackageUpdate(cmd gitCommander, name string) (bool, error) {
	path := filepath.Join(cmd.getConf().PackagePath(), name)
	if *cmd.verboseOpts().verboseFlg() {
		fmt.Fprintf(cmd.errs(), "[Info] Checking %s ...\n", name)
	}
	link, err := os.Readlink(path)
	if link != "" {
		if err != nil {
			// Just in case
			fmt.Fprintf(cmd.errs(), "Error! Reading link failed. Path = %s\n", path)
		}
		if *cmd.verboseOpts().verboseFlg() {
			fmt.Fprintln(cmd.errs(), "[Info] Symbolic link. Skip")
		}
		return false, nil
	}

	if err = os.Chdir(path); err != nil {
		fmt.Fprintf(cmd.errs(), "Error! Directory change failed. Path = %s\n", path)
		return false, ErrOperationFailed
	}
	return cmd.gitCtl().HasUpdate(*cmd.verboseOpts().verboseFlg())
}
