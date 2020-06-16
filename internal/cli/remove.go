package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"text/template"
)

type removeCmd struct {
	verboseCmd
	command string
}

func newRemoveCmd(common commonCmd) removeCmd {
	cmd := &removeCmd{}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "remove", nil)
	return *cmd
}

func (cmd *removeCmd) usage() {
	const help = `Summary:
  Uninstall a package.

Syntax:
  {{.Prog}} {{.Cmd}} <package>

Examples:
  {{.Prog}} {{.Cmd}} bats-core
  {{.Prog}} {{.Cmd}} enhancd

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, struct{ Prog, Cmd string }{cmd.name, cmd.command})

	cmd.flags.PrintDefaults()
}

func (cmd *removeCmd) parseAndExec(args []string) error {
	cmd.command = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStart(cmd, args[1:], true)
	if done || err != nil {
		return err
	}

	return removePackage(cmd, cmd.flags.Arg(0))
}

func removePackage(cmd verboseRunner, name string) error {
	path := filepath.Join(cmd.getConfig().PackagePath(), name)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.getErrs(), "\"%s\" is not installed\n", name)
		return ErrArgument
	}

	if err := removeBinsLinks(cmd, path); err != nil {
		return ErrOperationFailed
	}

	if err := os.RemoveAll(path); err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! Package removal failed. Path = %s\n", path)
		return ErrOperationFailed
	}

	fmt.Fprintf(cmd.getOuts(), "\"%s\" is removed\n", name)
	return nil
}

func removeBinsLinks(cmd verboseRunner, pkgPath string) error {
	binPath := cmd.getConfig().BinPath()
	bins, err := ioutil.ReadDir(binPath)
	if err != nil {
		fmt.Fprintf(cmd.getErrs(), "Error! %s\n", err)
		return err
	}

	for _, bin := range bins {
		sym := filepath.Join(binPath, bin.Name())
		src, err := os.Readlink(sym)
		if err != nil {
			fmt.Fprintf(cmd.getErrs(), "Warning! Failed to read link of %s. Error = %s\n", sym, err)
			continue
		}
		if strings.HasPrefix(src, pkgPath) {
			if *cmd.getVerboseOpts().getVerbose() {
				fmt.Fprintf(cmd.getOuts(), "Delete %s -> %s\n", sym, src)
			}
			if err = os.Remove(sym); err != nil {
				fmt.Fprintf(cmd.getErrs(), "Error! Deletion failed: %s. Error = %s\n", sym, err)
				return err
			}
		}
	}

	return nil
}
