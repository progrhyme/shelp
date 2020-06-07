package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/pflag"
)

type removeCmd struct {
	commonCmd
	name   string
	option struct {
		commonFlags
	}
}

func newRemoveCmd(common commonCmd) removeCmd {
	cmd := &removeCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("remove", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	return *cmd
}

func (cmd *removeCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Uninstall a package.

Syntax:
  %s %s <package>

Examples:
  %s %s bats-core
  %s %s enhancd

Options:
`, cmd.command, cmd.name, cmd.command, cmd.name, cmd.command, cmd.name)
	cmd.flags.PrintDefaults()
}

func (cmd *removeCmd) parseAndExec(args []string) error {
	cmd.name = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args[1:], true)
	if done || err != nil {
		return err
	}

	pkg := cmd.flags.Arg(0)
	path := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "Error! Target package \"%s\" not found at %s\n", pkg, path)
		return ErrArgument
	}

	if err = os.RemoveAll(path); err != nil {
		fmt.Fprintf(cmd.err, "Error! Package removal failed. Path = %s\n", path)
		return ErrOperationFailed
	}

	fmt.Fprintf(cmd.out, "%s is removed\n", pkg)
	return nil
}
