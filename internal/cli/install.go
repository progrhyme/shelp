package cli

import (
	"fmt"
	"path/filepath"
	"regexp"

	"github.com/progrhyme/claft/internal/git"
	"github.com/spf13/pflag"
)

type installCmd struct {
	commonCmd
	git    git.Git
	flags  pflag.FlagSet
	option struct {
		verbose *bool
		commonFlags
	}
}

func newInstallCmd(common commonCmd, git git.Git) installCmd {
	cmd := &installCmd{
		git:   git,
		flags: *pflag.NewFlagSet("install", pflag.ContinueOnError),
	}
	cmd.commonCmd = common

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *installCmd) usage() {
	fmt.Fprintf(cmd.err, `Install a GitHub repository as a package.

Syntax:
  %s install <account>/<repository>

Examples:
  %s install bats-core/bats-core
  %s install b4b4r07/enhancd

Options:
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
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		cmd.flags.Usage()
		return ErrParseFailed
	}

	if *cmd.option.help {
		cmd.flags.Usage()
		return nil
	}

	if cmd.flags.NArg() == 0 {
		cmd.flags.Usage()
		return ErrUsage
	}

	re := regexp.MustCompile(`\w+/\w+`)
	if !re.MatchString(cmd.flags.Arg(0)) {
		fmt.Fprintf(
			cmd.err,
			"Error! Given argument \"%s\" does not look like valid repository\n",
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
