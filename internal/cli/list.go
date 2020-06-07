package cli

import (
	"fmt"
	"io/ioutil"
	"os"

	"github.com/spf13/pflag"
)

type listCmd struct {
	commonCmd
	option struct {
		verbose *bool
		commonFlags
	}
}

func newListCmd(common commonCmd) listCmd {
	cmd := &listCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("list", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "# Verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
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
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, false)
	if done || err != nil {
		return err
	}

	nopkg := func() {
		fmt.Fprintln(cmd.err, "No package is installed")
	}
	if _, err := os.Stat(cmd.config.PackagePath()); os.IsNotExist(err) {
		nopkg()
		return nil
	}

	pkgs, err := ioutil.ReadDir(cmd.config.PackagePath())
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	if len(pkgs) == 0 {
		nopkg()
	} else {
		for _, pkg := range pkgs {
			fmt.Fprintln(cmd.out, pkg.Name())
		}
	}

	return nil
}
