package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh/terminal"
)

type destroyCmd struct {
	commonCmd
	option struct {
		yes *bool
		commonFlags
	}
}

func newDestroyCmd(common commonCmd) destroyCmd {
	cmd := &destroyCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("destroy", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.option.yes = cmd.flags.BoolP("yes", "y", false, "# Destroy without confirmation")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *destroyCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Delete all contents in %s including the root directory.

Syntax:
  %s destroy [-y|--yes]

Options:
`, config.RootVarName, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *destroyCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, false)
	if done || err != nil {
		return err
	}

	root := cmd.config.RootPath()
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "Not exist: %s\n", root)
		return ErrOperationFailed
	}

	if terminal.IsTerminal(0) {
		if !*cmd.option.yes {
			fmt.Fprintf(cmd.out, `Delete all contents in %s including packages and the directory itself.
Are you sure? (y/N) `, root)
			var ans string
			fmt.Scan(&ans)
			if !strings.HasPrefix(ans, "y") && !strings.HasPrefix(ans, "Y") {
				fmt.Fprintln(cmd.out, "Canceled")
				return ErrCanceled
			}
		}
	} else if !*cmd.option.yes {
		fmt.Fprintln(cmd.err, "Warning! Destruction is canceled because flag \"yes\" is not set")
		return ErrOperationFailed
	}

	if err = os.RemoveAll(root); err != nil {
		fmt.Fprintf(cmd.err, "Error! Destruction failed. Error = %v\n", err)
		return ErrOperationFailed
	}

	fmt.Fprintf(cmd.out, "Deleted: %s\n", root)
	return nil
}
