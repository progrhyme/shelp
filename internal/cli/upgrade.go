package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

type upgradeCmd struct {
	verboseCmd
	git git.Git
}

func newUpgradeCmd(common commonCmd, git git.Git) upgradeCmd {
	cmd := &upgradeCmd{git: git}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("upgrade", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "# Verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *upgradeCmd) usage() {
	const help = `Summary:
  Upgrade an installed package.

Syntax:
  {{.Prog}} {{.Cmd}} <package>

Examples:
  {{.Prog}} {{.Cmd}} bats-core
  {{.Prog}} {{.Cmd}} enhancd

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "upgrade"})

	cmd.flags.PrintDefaults()
}

func (cmd *upgradeCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(cmd, args, true)
	if done || err != nil {
		return err
	}

	pkg := cmd.flags.Arg(0)
	path := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is not installed\n", pkg)
		return ErrArgument
	}

	if err = os.Chdir(path); err != nil {
		fmt.Fprintf(cmd.err, "Error! Directory change failed. Path = %s\n", path)
		return ErrOperationFailed
	}

	err = cmd.git.Pull(*cmd.option.verbose)
	if err != nil {
		return ErrCommandFailed
	}
	return nil
}
