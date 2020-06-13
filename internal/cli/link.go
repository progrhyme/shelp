package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"text/template"
)

type linkCmd struct {
	verboseCmd
}

func newLinkCmd(common commonCmd) linkCmd {
	cmd := &linkCmd{}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "link", cmd.usage)
	return *cmd
}

func (cmd *linkCmd) usage() {
	const help = `Summary:
  Pseudo installation of a package from local filesystem.
  Creates symbolic link of a directory into a package path.

Syntax:
  {{.Prog}} {{.Cmd}} path/to/dir [<package-name>]

If you ommit "<package-name>" argument, the basename of the directory is used as the package name.

Examples:
  {{.Prog}} {{.Cmd}} .                   # Link current directory
  {{.Prog}} {{.Cmd}} path/to/foo-sh foo  # Link as package "foo"

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "link"})

	cmd.flags.PrintDefaults()
}

func (cmd *linkCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(cmd, args, true)
	if done || err != nil {
		return err
	}

	src := cmd.flags.Arg(0)
	if _, err := os.Stat(src); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "Error! \"%s\" does not exist\n", src)
		return ErrArgument
	}
	path, err := filepath.Abs(src)
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! Can't resolve path of \"%s\"\n", src)
		return ErrArgument
	}
	base := filepath.Base(path)

	if err = prepareInstallDirectories(cmd.config); err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	var pkg string
	if cmd.flags.NArg() > 1 {
		re := regexp.MustCompile(`^\w+`)
		if !re.MatchString(cmd.flags.Arg(1)) {
			fmt.Fprintf(
				cmd.err,
				"Error! Given argument \"%s\" does not look like valid package name\n",
				cmd.flags.Arg(1))
			return ErrArgument
		}
		pkg = cmd.flags.Arg(1)
	} else {
		pkg = base
	}
	pkgPath := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(pkgPath); !os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is already installed\n", pkg)
		return ErrArgument
	}

	if err = os.Symlink(path, pkgPath); err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	binPath := filepath.Join(pkgPath, "bin")
	var linkErr error
	if _, err := os.Stat(binPath); err == nil {
		linkErr = createBinsLinks(cmd, binPath)
	} else {
		linkErr = createBinsLinks(cmd, pkgPath)
	}
	if linkErr != nil {
		fmt.Fprintf(cmd.err, "\"%s\" is linked as package \"%s\", but with some failures\n", src, pkg)
		return linkErr
	}

	fmt.Fprintf(cmd.out, "\"%s\" is linked as package \"%s\"\n", src, pkg)
	return nil
}
