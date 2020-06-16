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
	t.Execute(cmd.errs, struct{ Prog, Cmd string }{cmd.name, "outdated"})

	cmd.flags.PrintDefaults()
}

func (cmd *outdatedCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
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
			fmt.Fprintln(cmd.outs, pkg.Name())
		} else {
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.errs, "%s is up-to-date\n", pkg.Name())
			}
		}
	}

	return nil
}

func hasPackageUpdate(cmd gitRunner, name string) (bool, error) {
	path := filepath.Join(cmd.getConfig().PackagePath(), name)
	if *cmd.getVerboseOpts().getVerbose() {
		fmt.Fprintf(cmd.getErrs(), "[Info] Checking %s ...\n", name)
	}
	if isSymlink(path, cmd.getErrs()) {
		if *cmd.getVerboseOpts().getVerbose() {
			fmt.Fprintln(cmd.getErrs(), "[Info] Symbolic link. Skip")
		}
		return false, nil
	}

	if err := os.Chdir(path); err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! Directory change failed. Path = %s\n", path)
		return false, ErrOperationFailed
	}
	return cmd.getGit().HasUpdate(*cmd.getVerboseOpts().getVerbose())
}
