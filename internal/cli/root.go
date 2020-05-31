package cli

import (
	"fmt"
	"io"

	"github.com/spf13/pflag"
)

type rootCmd struct {
	flags   pflag.FlagSet
	output  io.Writer
	command string
	version string
	option  struct {
		help    *bool
		version *bool
	}
}

func newRootCmd(out io.Writer, prog string, ver string) rootCmd {
	cmd := &rootCmd{
		flags:   *pflag.NewFlagSet("main", pflag.ContinueOnError),
		output:  out,
		command: prog,
		version: ver,
	}

	cmd.flags.SetOutput(out)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "show help")
	cmd.option.version = cmd.flags.BoolP("version", "v", false, "show version")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *rootCmd) usage() {
	fmt.Fprintf(cmd.output, `"%s" is a Git-based package manager for shell scripts written in Go.

Usage:
  %s -h|--help
  %s -v|--version
  %s COMMAND [arguments...] [option...]

Available subcommands:
  install  # install a package

`, cmd.command, cmd.command, cmd.command, cmd.command)
	fmt.Fprint(cmd.output, "option without subcommand:\n")
	cmd.flags.PrintDefaults()
}

func (cmd *rootCmd) parseAndExec(args []string) error {
	err := cmd.flags.Parse(args[1:])
	if err != nil {
		fmt.Fprintf(cmd.output, "Error! %s\n", err)
		cmd.flags.Usage()
		return ErrParseFailed
	}

	if cmd.withNoArg() {
		return nil
	}

	fmt.Fprintf(cmd.output, "Error! Subcommand not found: %s\n", cmd.flags.Arg(0))
	cmd.flags.Usage()
	return ErrUsage
}

func (cmd *rootCmd) withNoArg() bool {
	if *cmd.option.help {
		cmd.flags.Usage()
		return true
	}

	if *cmd.option.version {
		fmt.Printf("Version: %s\n", cmd.version)
		return true
	}

	return false
}
