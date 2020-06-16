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
	command string
}

type installArgs struct {
	from string
	as   string
	at   string
	bin  []string
}

// shelp package params
type shelpkg struct {
	name   string
	url    string
	branch string
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
	t.Execute(cmd.errs, struct{ Prog, Cmd string }{cmd.name, cmd.command})
	cmd.flags.PrintDefaults()
	fmt.Fprintf(cmd.errs, `
Limitation:
  1. This command always clones repository as shallow one, with "--depth=1" option
  2. You can't specify "--branch" option in the latter command syntax, nor others
`)
}

func (cmd *installCmd) parseAndExec(args []string) error {
	cmd.command = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStart(cmd, args[1:], true)
	if done || err != nil {
		return err
	}

	if err = prepareInstallDirectories(cmd.config); err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return ErrOperationFailed
	}

	param := installArgs{from: cmd.flags.Arg(0)}
	if cmd.flags.NArg() > 1 {
		param.as = cmd.flags.Arg(1)
	}
	return installPackage(cmd, param)
}

func prepareInstallDirectories(cfg *config.Config) error {
	if err := os.MkdirAll(cfg.PackagePath(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(cfg.BinPath(), 0755)
}

func packageToInstall(cmd verboseRunner, args installArgs) (shelpkg, error) {
	pkg := shelpkg{}
	var re *regexp.Regexp

	if args.as != "" {
		re = regexp.MustCompile(`^\w+`)
		if !re.MatchString(args.as) {
			fmt.Fprintf(
				cmd.getErrs(),
				"Error! Given argument \"%s\" does not look like valid package name\n",
				args.as)
			return pkg, ErrArgument
		}
		pkg.name = args.as
	}

	re = regexp.MustCompile(`^(?:([\w\-\.]+)/)?([\w\-\.]+)/([\w\-\.]+)(?:@([\w\-\.]+))?$`)
	if re.MatchString(args.from) {
		matched := re.FindStringSubmatch(args.from)
		site := matched[1]
		if site == "" {
			site = "github.com"
		}
		account := matched[2]
		repo := matched[3]
		pkg.branch = matched[4]
		pkg.url = fmt.Sprintf("https://%s/%s/%s.git", site, account, repo)
		if pkg.name == "" {
			pkg.name = repo
		}
	} else {
		pkg.url = args.from
		if pkg.name == "" {
			pkg.name = strings.TrimSuffix(filepath.Base(pkg.url), ".git")
		}
	}

	if args.at != "" {
		pkg.branch = args.at
	}

	return pkg, nil
}

func installPackage(cmd gitRunner, args installArgs) error {
	pkg, err := packageToInstall(cmd, args)
	if err != nil {
		return err
	}

	pkgPath := filepath.Join(cmd.getConfig().PackagePath(), pkg.name)
	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		fmt.Fprintf(cmd.getErrs(), "\"%s\" is already installed\n", pkg.name)
		return ErrAlreadyInstalled
	}

	fmt.Fprintf(cmd.getOuts(), "Fetching \"%s\" from %s ...\n", pkg.name, pkg.url)
	gitOpts := git.Option{
		Branch:  pkg.branch,
		Shallow: cmd.getConfig().Git.Shallow,
		Verbose: *cmd.getVerboseOpts().getVerbose(),
	}
	err = cmd.getGit().Clone(pkg.url, pkgPath, gitOpts)
	if err != nil {
		return ErrCommandFailed
	}

	var linkErr error
	if len(args.bin) > 0 {
		for _, bin := range args.bin {
			linkErr = createLinkByBinAndDir(cmd, bin, pkgPath)
		}
	} else {
		binPath := filepath.Join(pkgPath, "bin")
		if _, err := os.Stat(binPath); err == nil {
			linkErr = createLinksByBinDir(cmd, binPath)
		} else {
			linkErr = createLinksByBinDir(cmd, pkgPath)
		}
	}
	if linkErr != nil {
		fmt.Fprintf(cmd.getErrs(), "\"%s\" is installed, but with some failures\n", pkg.name)
		return linkErr
	}

	fmt.Fprintf(cmd.getOuts(), "\"%s\" is successfully installed\n", pkg.name)
	return nil
}

func createLinksByBinDir(cmd verboseRunner, path string) error {
	files, err := ioutil.ReadDir(path)
	if err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! %s\n", err)
		return ErrOperationFailed
	}

	var warn bool
	for _, file := range files {
		if !file.IsDir() && isExecutable(file.Mode()) {
			err = createLinkByBinAndDir(cmd, file.Name(), path)
			switch err {
			case ErrWarning:
				warn = true
				continue
			case nil:
				// Nothing to do
			default:
				return err
			}
		}
	}
	if warn {
		return ErrWarning
	}

	return nil
}

func createLinkByBinAndDir(cmd verboseRunner, bin, path string) error {
	exe := filepath.Join(path, bin)
	sym := filepath.Join(cmd.getConfig().BinPath(), filepath.Base(bin))
	if *cmd.getVerboseOpts().getVerbose() {
		fmt.Fprintf(cmd.getOuts(), "Symlink: %s -> %s\n", sym, exe)
	}
	if _, err := os.Stat(sym); !os.IsNotExist(err) {
		fmt.Fprintf(cmd.getErrs(), "Warning! Can't create link of %s which already exists\n", exe)
		return ErrWarning
	}
	if err := os.Symlink(exe, sym); err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! %s\n", err)
		return ErrOperationFailed
	}
	return nil
}

func isExecutable(mode os.FileMode) bool {
	return mode&0111 != 0
}
