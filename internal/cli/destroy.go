package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/progrhyme/shelp/internal/config"
	"golang.org/x/crypto/ssh/terminal"
)

type destroyCmd struct {
	commonCmd
	option struct {
		yes *bool
		commonFlags
	}
}

func (cmd *destroyCmd) getOpts() flagger {
	return &cmd.option
}

func newDestroyCmd(common commonCmd) destroyCmd {
	cmd := &destroyCmd{}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "destroy", cmd.usage)
	cmd.option.yes = cmd.flags.BoolP("yes", "y", false, "# Destroy without confirmation")
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
	done, err := parseStart(cmd, args, false)
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
