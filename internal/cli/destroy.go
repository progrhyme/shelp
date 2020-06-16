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
		commonOpts
	}
}

func (cmd *destroyCmd) getOpts() flavor {
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
	fmt.Fprintf(cmd.errs, `Summary:
  Delete all contents in %s including the root directory.

Syntax:
  %s destroy [-y|--yes]

Options:
`, config.RootVarName, cmd.name)
	cmd.flags.PrintDefaults()
}

func (cmd *destroyCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
	if done || err != nil {
		return err
	}

	root := cmd.config.RootPath()
	if _, err := os.Stat(root); os.IsNotExist(err) {
		fmt.Fprintf(cmd.errs, "Not exist: %s\n", root)
		return ErrOperationFailed
	}

	if terminal.IsTerminal(0) {
		if !*cmd.option.yes {
			fmt.Fprintf(cmd.outs, `Delete all contents in %s including packages and the directory itself.
Are you sure? (y/N) `, root)
			var ans string
			fmt.Scan(&ans)
			if !strings.HasPrefix(ans, "y") && !strings.HasPrefix(ans, "Y") {
				fmt.Fprintln(cmd.outs, "Canceled")
				return ErrCanceled
			}
		}
	} else if !*cmd.option.yes {
		fmt.Fprintln(cmd.errs, "Warning! Destruction is canceled because flag \"yes\" is not set")
		return ErrOperationFailed
	}

	if err = os.RemoveAll(root); err != nil {
		fmt.Fprintf(cmd.errs, "Error! Destruction failed. Error = %v\n", err)
		return ErrOperationFailed
	}

	fmt.Fprintf(cmd.outs, "Deleted: %s\n", root)
	return nil
}
