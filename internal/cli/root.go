package cli

import (
	"fmt"
	"text/template"

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
	cmd.option.version = cmd.flags.BoolP("version", "v", false, "# Show CLI version")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *rootCmd) usage() {
	const help = `Summary:
  "{{.Prog}}" is a Git-based package manager for shell scripts written in Go.

Usage:
  {{.Prog}} COMMAND [arguments...] [options...]
  {{.Prog}} -h|--help
  {{.Prog}} -v|--version

Available Commands:
  init       # Initialize {{.Prog}} for shell environment
  install    # Install a package
  add        # Alias of "install"
  remove     # Uninstall a package
  uninstall  # Alias of "remove"
  list       # List installed packages
  destroy    # Delete all materials including packages

Run "{{.Prog}} COMMAND -h|--help" to see usage of each command.

Options without subcommand:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog string }{cmd.command})

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
