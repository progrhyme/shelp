package cli

import (
	"fmt"
	"io"
	"path/filepath"
	"regexp"

	"github.com/progrhyme/claft/internal/config"
	"github.com/progrhyme/claft/internal/git"
	"github.com/spf13/pflag"
)

type installCmd struct {
	flags   pflag.FlagSet
	output  io.Writer
	config  config.Config
	git     git.Git
	command string
	option  struct {
		verbose *bool
	}
}

func newInstallCmd(out io.Writer, cfg config.Config, git git.Git, prog string) installCmd {
	cmd := &installCmd{
		flags:   *pflag.NewFlagSet("install", pflag.ContinueOnError),
		output:  out,
		config:  cfg,
		git:     git,
		command: prog,
	}

	cmd.flags.SetOutput(out)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "verbose output")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *installCmd) usage() {
	fmt.Fprintf(cmd.output, `Syntax:
  %s install <account>/<repository>

Examples:
  %s install bats-core/bats-core
  %s install b4b4r07/enhancd

option:
`, cmd.command, cmd.command, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *installCmd) parseAndExec(args []string) error {
	if len(args) == 0 {
		cmd.flags.Usage()
		return ErrUsage
	}

	err := cmd.flags.Parse(args)
	if err != nil {
		fmt.Fprintf(cmd.output, "Error! %s\n", err)
		cmd.flags.Usage()
		return ErrParseFailed
	}

	if cmd.flags.NArg() == 0 {
		cmd.flags.Usage()
		return ErrUsage
	}

	re := regexp.MustCompile(`\w+/\w+`)
	if !re.MatchString(cmd.flags.Arg(0)) {
		fmt.Fprintf(
			cmd.output,
			"Error! Given argument %s does not look like valid repository\n",
			cmd.flags.Arg(0))
		cmd.flags.Usage()
		return ErrArgument
	}

	pkg := filepath.Base(cmd.flags.Arg(0))
	url := fmt.Sprintf("https://github.com/%s.git", cmd.flags.Arg(0))
	err = cmd.git.Clone(url, filepath.Join(cmd.config.PackagePath(), pkg))
	if err != nil {
		return ErrCommandFailed
	}

	return nil
}
