package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type removeCmd struct {
	commonCmd
	flags  pflag.FlagSet
	option struct {
		commonFlags
	}
}

func newRemoveCmd(common commonCmd) removeCmd {
	cmd := &removeCmd{
		flags: *pflag.NewFlagSet("remove", pflag.ContinueOnError),
	}
	cmd.commonCmd = common

	cmd.flags.SetOutput(cmd.err)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *removeCmd) usage() {
	fmt.Fprintf(cmd.err, `Uninstall a package.

Syntax:
  %s remove <package>

Examples:
  %s remove bats-core
  %s remove enhancd

Options:
`, cmd.command, cmd.command, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *removeCmd) parseAndExec(args []string) error {
	if len(args) == 0 {
		cmd.flags.Usage()
		return ErrUsage
	}

	err := cmd.flags.Parse(args)
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		cmd.flags.Usage()
		return ErrParseFailed
	}

	if *cmd.option.help {
		cmd.flags.Usage()
		return nil
	}

	if cmd.flags.NArg() == 0 {
		cmd.flags.Usage()
		return ErrUsage
	}

	pkg := cmd.flags.Arg(0)
	path := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "Error! Target package \"%s\" not found at %s\n", pkg, path)
		return ErrArgument
	}

	if err = os.RemoveAll(path); err != nil {
		fmt.Fprintf(cmd.err, "Error! Package removal failed. Path = %s\n", path)
		return ErrCommandFailed
	}

	fmt.Fprintf(cmd.out, "%s is removed\n", pkg)
	return nil
}
