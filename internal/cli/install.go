package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"text/template"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
)

type installCmd struct {
	gitCmd
	name string
}

func newInstallCmd(common commonCmd, git git.Git) installCmd {
	cmd := &installCmd{}
	cmd.commonCmd = common
	cmd.git = git
	setupCmdFlags(cmd, "install", nil)
	return *cmd
}

func (cmd *installCmd) usage() {
	const help = `Summary:
  Install a repository from HTTPS site as a {{.Prog}} package, assuming it contains shell scripts.

Syntax:
  # Handy syntax using HTTPS protocol
  {{.Prog}} {{.Cmd}} [<site>/]<account>/<repository>[@<branch>] [<package-name>]

  # Specify complete git-url with any protocol
  {{.Prog}} {{.Cmd}} <git-url> [<package-name>]

If you ommit preceding "<site>/" specifier in former syntax, "github.com" is used by default.

Examples:
  # Handy syntax
  {{.Prog}} {{.Cmd}} b4b4r07/enhancd           # Install "enhancd" from github.com
  {{.Prog}} {{.Cmd}} b4b4r07/enhancd@v2.2.4    # Install specified tag or branch
  {{.Prog}} {{.Cmd}} bats-core/bats-core bats  # Install as "bats"
  {{.Prog}} {{.Cmd}} gitlab.com/dwt1/dotfiles  # Install from gitlab.com

  # Specify git-url
  {{.Prog}} {{.Cmd}} git@github.com:b4b4r07/enhancd.git  # Install via SSH protocol
  {{.Prog}} {{.Cmd}} file:///path/to/repository          # Install via Local protocol
  {{.Prog}} {{.Cmd}} git://server/gitproject.git         # Install via Git protocol

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, cmd.name})
	cmd.flags.PrintDefaults()
	fmt.Fprintf(cmd.err, `
Limitation:
  1. This command always clones repository as shallow one, with "--depth=1" option
  2. You can't specify "--branch" option in the latter command syntax, nor others
`)
}

func (cmd *installCmd) parseAndExec(args []string) error {
	cmd.name = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStartHelp(cmd, args[1:], true)
	if done || err != nil {
		return err
	}

	if err = prepareInstallDirectories(cmd.config); err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	var pkg, url, branch string
	var re *regexp.Regexp

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
	}

	re = regexp.MustCompile(`^(?:([\w\-\.]+)/)?([\w\-\.]+)/([\w\-\.]+)(?:@([\w\-\.]+))?$`)
	if re.MatchString(cmd.flags.Arg(0)) {
		matched := re.FindStringSubmatch(cmd.flags.Arg(0))
		site := matched[1]
		if site == "" {
			site = "github.com"
		}
		account := matched[2]
		repo := matched[3]
		branch = matched[4]
		url = fmt.Sprintf("https://%s/%s/%s.git", site, account, repo)
		if pkg == "" {
			pkg = repo
		}
	} else {
		url = cmd.flags.Arg(0)
		if pkg == "" {
			pkg = strings.TrimSuffix(filepath.Base(url), ".git")
		}
	}

	pkgPath := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is already installed\n", pkg)
		return ErrArgument
	}

	fmt.Fprintf(cmd.out, "Fetching \"%s\" from %s ...\n", pkg, url)
	err = cmd.git.Clone(url, pkgPath, branch, *cmd.option.verbose)
	if err != nil {
		return ErrCommandFailed
	}

	binPath := filepath.Join(pkgPath, "bin")
	var linkErr error
	if _, err := os.Stat(binPath); err == nil {
		linkErr = createBinsLinks(cmd, binPath)
	} else {
		linkErr = createBinsLinks(cmd, pkgPath)
	}
	if linkErr != nil {
		fmt.Fprintf(cmd.err, "\"%s\" is installed, but with some failures\n", pkg)
		return linkErr
	}

	fmt.Fprintf(cmd.out, "\"%s\" is successfully installed\n", pkg)
	return nil
}

func prepareInstallDirectories(cfg config.Config) error {
	if err := os.MkdirAll(cfg.PackagePath(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(cfg.BinPath(), 0755)
}

func createBinsLinks(cmd verboseCommander, path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(cmd.errs(), "Error! %s\n", err)
		return ErrOperationFailed
	}

	var warn bool
	for _, file := range files {
		if !file.IsDir() && isExecutable(file.Mode()) {
			exe := filepath.Join(path, file.Name())
			sym := filepath.Join(cmd.getConf().BinPath(), file.Name())
			if *cmd.verboseOpts().verboseFlg() {
				fmt.Fprintf(cmd.outs(), "Symlink: %s -> %s\n", sym, exe)
			}
			if _, err := os.Stat(sym); !os.IsNotExist(err) {
				fmt.Fprintf(cmd.errs(), "Warning! Can't create link of %s which already exists\n", exe)
				warn = true
				continue
			}
			if err = os.Symlink(exe, sym); err != nil {
				fmt.Fprintf(cmd.errs(), "Error! %s\n", err)
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
