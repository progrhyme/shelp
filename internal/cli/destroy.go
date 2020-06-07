package cli

import (
	"fmt"
	"os"

	"github.com/progrhyme/claft/internal/config"
	"github.com/spf13/pflag"
)

type destroyCmd struct {
	commonCmd
	option struct {
		commonFlags
	}
}

func newDestroyCmd(common commonCmd) destroyCmd {
	cmd := &destroyCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("destroy", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *destroyCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Delete all contents in %s including the root directory.

Syntax:
  %s destroy

Options:
`, config.RootVarName, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *destroyCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, false)
	if done || err != nil {
		return err
	}

	if err = os.RemoveAll(cmd.config.RootPath()); err != nil {
		fmt.Fprintf(cmd.err, "Error! Destruction failed. Error = %v\n", err)
		return ErrCommandFailed
	}

	fmt.Fprintf(cmd.out, "Deleted: %s\n", cmd.config.RootPath())
	return nil
}
