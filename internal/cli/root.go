package cli

import (
	"fmt"

	"github.com/spf13/pflag"
)

type rootCmd struct {
	commonCmd
	version string
	option  struct {
		version *bool
		commonFlags
	}
}

func newRootCmd(common commonCmd, ver string) rootCmd {
	cmd := &rootCmd{version: ver}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("main", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.version = cmd.flags.BoolP("version", "v", false, "show version")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *rootCmd) usage() {
	fmt.Fprintf(cmd.err, `"%s" is a Git-based package manager for shell scripts written in Go.

Usage:
  %s -h|--help
  %s -v|--version
  %s COMMAND [arguments...] [option...]

Available subcommands:
  install  # install a package
  remove   # uninstall a package
  list     # list installed packages

`, cmd.command, cmd.command, cmd.command, cmd.command)
	fmt.Fprint(cmd.err, "Options without subcommand:\n")
	cmd.flags.PrintDefaults()
}

func (cmd *rootCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, false)
	if done || err != nil {
		return err
	}

	if *cmd.option.version {
		fmt.Fprintf(cmd.out, "Version: %s\n", cmd.version)
		return nil
	}

	fmt.Fprintf(cmd.err, "Error! Subcommand not found: %s\n", cmd.flags.Arg(0))
	cmd.flags.Usage()
	return ErrUsage
}
