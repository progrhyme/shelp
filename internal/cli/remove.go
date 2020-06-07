package cli

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/pflag"
)

type removeCmd struct {
	commonCmd
	name   string
	option struct {
		verbose *bool
		commonFlags
	}
}

func newRemoveCmd(common commonCmd) removeCmd {
	cmd := &removeCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("remove", pflag.ContinueOnError)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.verbose = cmd.flags.BoolP("verbose", "v", false, "# Verbose output")
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	return *cmd
}

func (cmd *removeCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Uninstall a package.

Syntax:
  %s %s <package>

Examples:
  %s %s bats-core
  %s %s enhancd

Options:
`, cmd.command, cmd.name, cmd.command, cmd.name, cmd.command, cmd.name)
	cmd.flags.PrintDefaults()
}

func (cmd *removeCmd) parseAndExec(args []string) error {
	cmd.name = args[0]
	cmd.flags.Usage = cmd.usage

	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args[1:], true)
	if done || err != nil {
		return err
	}

	pkg := cmd.flags.Arg(0)
	path := filepath.Join(cmd.config.PackagePath(), pkg)
	if _, err := os.Stat(path); os.IsNotExist(err) {
		fmt.Fprintf(cmd.err, "\"%s\" is not installed\n", pkg)
		return ErrArgument
	}

	if err = cmd.removeBinsLinks(path); err != nil {
		return ErrOperationFailed
	}

	if err = os.RemoveAll(path); err != nil {
		fmt.Fprintf(cmd.err, "Error! Package removal failed. Path = %s\n", path)
		return ErrOperationFailed
	}

	fmt.Fprintf(cmd.out, "\"%s\" is removed\n", pkg)
	return nil
}

func (cmd *removeCmd) removeBinsLinks(pkgPath string) error {
	binPath := cmd.config.BinPath()
	bins, err := ioutil.ReadDir(binPath)
	if err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return err
	}

	for _, bin := range bins {
		sym := filepath.Join(binPath, bin.Name())
		src, err := os.Readlink(sym)
		if err != nil {
			fmt.Fprintf(cmd.err, "Warning! Failed to read link of %s. Error = %s\n", src, err)
			continue
		}
		if strings.HasPrefix(src, pkgPath) {
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.out, "Delete %s -> %s\n", sym, src)
			}
			if err = os.Remove(sym); err != nil {
				fmt.Fprintf(cmd.err, "Error! Deletion failed: %s. Error = %s\n", sym, err)
			}
		}
	}

	return nil
}
