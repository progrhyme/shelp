package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

type outdatedCmd struct {
	verboseCmd
	git git.Git
}

func newOutdatedCmd(common commonCmd, git git.Git) outdatedCmd {
	cmd := &outdatedCmd{git: git}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("outdated", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "# Verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func (cmd *outdatedCmd) usage() {
	const help = `Summary:
  Show installed packages which can be updated.

Syntax:
  {{.Prog}} {{.Cmd}}

To update a package, run "{{.Prog}} upgrade <package>".

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "outdated"})

	cmd.flags.PrintDefaults()
}

func (cmd *outdatedCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(cmd, args, false)
	if done || err != nil {
		return err
	}

	nopkg := func() {
		fmt.Fprintln(cmd.err, "No package is installed")
	}
	if _, err := os.Stat(cmd.config.PackagePath()); os.IsNotExist(err) {
		nopkg()
		return nil
	}

	pkgs, err := ioutil.ReadDir(cmd.config.PackagePath())
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	if len(pkgs) == 0 {
		nopkg()
	} else {
		for _, pkg := range pkgs {
			path := filepath.Join(cmd.config.PackagePath(), pkg.Name())
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.err, "[Info] Checking %s ...\n", pkg.Name())
			}
			link, err := os.Readlink(path)
			if link != "" {
				if err != nil {
					fmt.Fprintf(cmd.err, "Error! Reading link failed. Path = %s\n", path)
				}
				if *cmd.option.verbose {
					fmt.Fprintln(cmd.err, "[Info] Symbolic link. Skip")
				}
				continue
			}

			if err = os.Chdir(path); err != nil {
				fmt.Fprintf(cmd.err, "Error! Directory change failed. Path = %s\n", path)
				return ErrOperationFailed
			}
			old, err := cmd.git.HasUpdate(*cmd.option.verbose)
			if err != nil {
				return ErrCommandFailed
			}
			if old {
				fmt.Fprintln(cmd.out, pkg.Name())
			} else {
				if *cmd.option.verbose {
					fmt.Fprintf(cmd.err, "%s is up-to-date\n", pkg.Name())
				}
			}
		}
	}

	return nil
}
