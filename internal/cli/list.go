package cli

import (
	"fmt"
)

type listCmd struct {
	verboseCmd
}

func newListCmd(common commonCmd) listCmd {
	cmd := &listCmd{}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "list", cmd.usage)
	return *cmd
}

func (cmd *listCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  List installed packages.

Syntax:
  %s list

Options:
`, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *listCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(cmd, args, false)
	if done || err != nil {
		return err
	}

	pkgs, err := installedPackages(cmd, false)
	if err != nil {
		return err
	}

	for _, pkg := range pkgs {
		fmt.Fprintln(cmd.out, pkg.Name())
	}

	return nil
}
