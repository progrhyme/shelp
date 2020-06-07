package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"

	"github.com/progrhyme/claft/internal/git"
	"github.com/spf13/pflag"
)

type installCmd struct {
	commonCmd
	git    git.Git
	option struct {
		verbose *bool
		commonFlags
	}
}

func newInstallCmd(common commonCmd, git git.Git) installCmd {
	cmd := &installCmd{git: git}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("install", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "# Verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *installCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Install a repository in github.com as a %s package, assuming it contains shell scripts.

Syntax:
  %s install <account>/<repository>

Examples:
  %s install bats-core/bats-core
  %s install b4b4r07/enhancd

Options:
`, cmd.command, cmd.command, cmd.command, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *installCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, true)
	if done || err != nil {
		return err
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

	if err = cmd.prepareInstallDirectories(); err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	pkg := filepath.Base(cmd.flags.Arg(0))
	pkgPath := filepath.Join(cmd.config.PackagePath(), pkg)
	url := fmt.Sprintf("https://github.com/%s.git", cmd.flags.Arg(0))
	err = cmd.git.Clone(url, pkgPath)
	if err != nil {
		return ErrCommandFailed
	}

	binPath := filepath.Join(pkgPath, "bin")
	if _, err := os.Stat(binPath); err == nil {
		err = cmd.createBinsLinks(binPath)
	} else {
		err = cmd.createBinsLinks(pkgPath)
	}
	if err != nil {
		return err
	}

	return nil
}

func (cmd *installCmd) prepareInstallDirectories() error {
	if err := os.MkdirAll(cmd.config.PackagePath(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(cmd.config.BinPath(), 0755)
}

func (cmd *installCmd) createBinsLinks(path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	for _, file := range files {
		if !file.IsDir() && isExecutable(file.Mode()) {
			exe := filepath.Join(path, file.Name())
			sym := filepath.Join(cmd.config.BinPath(), file.Name())
			fmt.Fprintf(cmd.out, "Symlink: %s -> %s\n", sym, exe)
			if err = os.Symlink(exe, sym); err != nil {
				fmt.Fprintf(cmd.err, "Error! %s\n", err)
				return ErrOperationFailed
			}
		}
	}

	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
