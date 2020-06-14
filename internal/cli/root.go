package cli

import (
	"fmt"
	"text/template"
)

type rootCmd struct {
	commonCmd
	version string
	option  struct {
		version *bool
		commonFlags
	}
}

func (cmd *rootCmd) getOpts() flagger {
	return &cmd.option
}

func newRootCmd(common commonCmd, ver string) rootCmd {
	cmd := &rootCmd{version: ver}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "main", cmd.usage)
	cmd.option.version = cmd.flags.BoolP("version", "v", false, "# Show CLI version")
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
  upgrade    # Upgrade installed packages
  outdated   # Show outdated packages
  link       # Pseudo installation of local directory
  destroy    # Delete all materials including packages

Run "{{.Prog}} COMMAND -h|--help" to see usage of each command.

Options without subcommand:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog string }{cmd.command})

	cmd.flags.PrintDefaults()
}

func (cmd *rootCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
	if done || err != nil {
		return err
	}

	if *cmd.option.version {
		fmt.Fprintf(cmd.out, "Version: %s\n", cmd.version)
		return nil
	}

	if cmd.flags.NArg() > 0 {
		fmt.Fprintf(cmd.err, "Error! Subcommand not found: %s\n", cmd.flags.Arg(0))
	}

	cmd.flags.Usage()
	return ErrUsage
}
