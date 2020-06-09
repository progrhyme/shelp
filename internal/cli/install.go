package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

type installCmd struct {
	commonCmd
	name   string
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
	return *cmd
}

func (cmd *installCmd) usage() {
	const help = `Summary:
  Install a repository in github.com as a {{.Prog}} package, assuming it contains shell scripts.

Syntax:
  {{.Prog}} {{.Cmd}} <account>/<repository> [<package-name>]

Examples:
  {{.Prog}} {{.Cmd}} bats-core/bats-core bats  # Install as "bats"
  {{.Prog}} {{.Cmd}} b4b4r07/enhancd           # Install as "enhancd"

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, cmd.name})

	cmd.flags.PrintDefaults()
}

func (cmd *installCmd) parseAndExec(args []string) error {
	cmd.name = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args[1:], true)
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

	var pkg string
	if cmd.flags.NArg() > 1 {
		re = regexp.MustCompile(`^\w+`)
		if !re.MatchString(cmd.flags.Arg(1)) {
			fmt.Fprintf(
				cmd.err,
				"Error! Given argument \"%s\" does not look like valid package name\n",
				cmd.flags.Arg(1))
			return ErrArgument
		}
		pkg = cmd.flags.Arg(1)
	} else {
		pkg = filepath.Base(cmd.flags.Arg(0))
	}
	pkgPath := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is already installed\n", pkg)
		return ErrArgument
	}

	url := fmt.Sprintf("https://github.com/%s.git", cmd.flags.Arg(0))
	fmt.Fprintf(cmd.out, "Fetching \"%s\" from %s ...\n", pkg, url)
	err = cmd.git.Clone(url, pkgPath, *cmd.option.verbose)
	if err != nil {
		return ErrCommandFailed
	}

	binPath := filepath.Join(pkgPath, "bin")
	var linkErr error
	if _, err := os.Stat(binPath); err == nil {
		linkErr = cmd.createBinsLinks(binPath)
	} else {
		linkErr = cmd.createBinsLinks(pkgPath)
	}
	if linkErr != nil {
		fmt.Fprintf(cmd.err, "\"%s\" is installed, but with some failures\n", pkg)
		return linkErr
	}

	fmt.Fprintf(cmd.out, "\"%s\" is successfully installed\n", pkg)
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

	var warn bool
	for _, file := range files {
		if !file.IsDir() && isExecutable(file.Mode()) {
			exe := filepath.Join(path, file.Name())
			sym := filepath.Join(cmd.config.BinPath(), file.Name())
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.out, "Symlink: %s -> %s\n", sym, exe)
			}
			if _, err := os.Stat(sym); !os.IsNotExist(err) {
				fmt.Fprintf(cmd.err, "Warning! Can't create link of %s which already exists\n", exe)
				warn = true
				continue
			}
			if err = os.Symlink(exe, sym); err != nil {
				fmt.Fprintf(cmd.err, "Error! %s\n", err)
				return ErrOperationFailed
			}
		}
	}
	if warn {
		return ErrWarning
	}

	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
