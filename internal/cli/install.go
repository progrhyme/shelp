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
	from      string
	as        string
	at        string
	bin       []string
	overwrite bool
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
  Install a git repository as a {{.Prog}} package and create symlinks for executable files in it.

Syntax:
  # Handy syntax using HTTPS protocol
  {{.Prog}} {{.Cmd}} [<site>/]<account>/<repository>[@<ref>] [<package-name>]

  # Specify complete git-url with any protocol
  {{.Prog}} {{.Cmd}} <git-url> [<package-name>]

If you ommit preceding "<site>/" specifier in former syntax, "github.com" is used by default.
You can specify any branch or tag or commit hash for "@<ref>" parameter.

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
  1. Unless you specify commit hash by "@<ref>" param, this command clones repository as shallow
     one by "--depth=1" option.
  2. You can't specify "--branch" option in the latter command syntax, nor others.
     Consider using "bundle" command with configuration file to do it.
`)
}

func (cmd *installCmd) parseAndExec(args []string) error {
	cmd.command = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStart(cmd, args[1:], true, false)
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
	if err := os.MkdirAll(cfg.BinPath(), 0755); err != nil {
		return err
	}
	return os.MkdirAll(cfg.TempPath(), 0755)
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
		pkg.ref = matched[4]
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
		pkg.ref = args.at
	}
	if pkg.ref != "" {
		re = regexp.MustCompile(`^[0-9a-f]{7,}$`)
		if re.MatchString(pkg.ref) {
			pkg.isCommitHash = true
		}
	}

	return pkg, nil
}

func packageInstalled(cmd gitRunner, path string) (shelpkg, error) {
	pkg := shelpkg{}
	pwd, err := chdir(cmd, path)
	if err != nil {
		return pkg, ErrOperationFailed
	}
	defer os.Chdir(pwd)

	repo, err := cmd.getGit().Worktree(*cmd.getVerboseOpts().getVerbose())
	if err != nil {
		return pkg, ErrOperationFailed
	}

	pkg.url = repo.RemoteURL
	pkg.ref = repo.BranchOrTag()
	if pkg.ref == repo.Branch && repo.IsBranchDefault() {
		pkg.isBranchDefault = true
	}

	return pkg, nil
}

func installPackage(cmd gitRunner, args installArgs) error {
	pkg, err := packageToInstall(cmd, args)
	if err != nil {
		return err
	}

	pkgPath := filepath.Join(cmd.getConfig().PackagePath(), pkg.name)
	reinstall := false

	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		alreadyInstalled := func() error {
			fmt.Fprintf(cmd.getErrs(), "\"%s\" is already installed\n", pkg.name)
			return ErrAlreadyInstalled
		}
		if !args.overwrite {
			return alreadyInstalled()
		}

		// Always overwrite pseudo package created by "link" command
		now := shelpkg{}
		if !isSymlink(pkgPath, cmd.getErrs()) {
			now, err = packageInstalled(cmd, pkgPath)
			if err != nil {
				return err
			}
			if pkg.isEquivalent(now) {
				return alreadyInstalled()
			}
		}

		// Re-install
		reinstall = true
		fmt.Fprintf(cmd.getErrs(), "Re-install \"%s\"\n", pkg.name)
		if pkg.url != now.url {
			fmt.Fprintf(cmd.getErrs(), "  from: %s => %s\n", now.url, pkg.url)
		}
		if pkg.ref != now.ref && (pkg.ref != "" || !now.isBranchDefault) {
			newref := pkg.ref
			if newref == "" {
				newref = "(default)"
			}
			fmt.Fprintf(cmd.getErrs(), "  at: %s => %s\n", now.ref, newref)
		}
	}

	gitOpts := git.Option{
		Shallow: cmd.getConfig().Git.Shallow,
		Verbose: *cmd.getVerboseOpts().getVerbose(),
	}
	if pkg.isCommitHash {
		gitOpts.Commit = pkg.ref
	} else {
		gitOpts.Branch = pkg.ref
	}
	tmpath := filepath.Join(cmd.getConfig().TempPath(), pkg.name)
	err = cmd.getGit().Clone(pkg.url, tmpath, gitOpts)

	defer func() {
		if _, err := os.Stat(tmpath); !os.IsNotExist(err) {
			if err := os.RemoveAll(tmpath); err != nil {
				fmt.Fprintf(cmd.getErrs(), "Error! Directory removal failed. Path = %s\n", tmpath)
			}
		}
	}()

	if err != nil {
		fmt.Fprintf(
			cmd.getErrs(), "Error! Installation failed. Package = %s, From = %s\n", pkg.name, pkg.url)
		return ErrCommandFailed
	}

	if reinstall {
		fmt.Fprintf(cmd.getOuts(), "Removing existing \"%s\" ... ", pkg.name)
		if err = removePackage(cmd, pkg.name, true); err != nil {
			return err
		}
		fmt.Fprintln(cmd.getOuts(), "Done")
	}

	if err = os.Rename(tmpath, pkgPath); err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! Moving directory failed: %s => %s\n", tmpath, pkgPath)
		return ErrOperationFailed
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
