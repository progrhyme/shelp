package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/progrhyme/claft/internal/config"
	"github.com/spf13/pflag"
)

type initCmd struct {
	commonCmd
	shell  string
	shProf string
	option struct {
		commonFlags
	}
}

func newInitCmd(common commonCmd) initCmd {
	cmd := &initCmd{}
	cmd.commonCmd = common
	cmd.flags = *pflag.NewFlagSet("init", pflag.ContinueOnError)

	cmd.shell = filepath.Base(os.Getenv("SHELL"))
	cmd.shProf = shellProfile(cmd.shell)

	cmd.flags.SetOutput(cmd.err)
	cmd.option.help = cmd.flags.BoolP("help", "h", false, "# Show help")
	cmd.flags.Usage = cmd.usage
	return *cmd
}

func shellProfile(shell string) string {
	var prof string
	switch shell {
	case "bash":
		home, err := os.UserHomeDir()
		if err != nil {
			panic(fmt.Sprintf("Can't get HomeDir! Error: %v", err))
		}
		if _, err := os.Stat(filepath.Join(home, ".bashrc")); !os.IsNotExist(err) {
			prof = "~/.bashrc"
		} else {
			prof = "~/.bash_profile"
		}
	case "zsh":
		prof = "~/.zshrc"
	default:
		prof = "its profile"
	}
	return prof
}

func (cmd *initCmd) usage() {
	fmt.Fprintf(cmd.err, `Summary:
  Enable %s in one's shell environment.

Usage:
    %s init - [SHELL]  # Print scripts (for specified SHELL)

  To enable %s automatically in one's shell, append the following to %s:

    eval "$(%s init -)"

  It prints scripts for current shell unless user specify SHELL argument.

Limitation:
  Only POSIX compatible shells are supported for now.

Options:
`, cmd.command, cmd.command, cmd.command, cmd.shProf, cmd.command)
	cmd.flags.PrintDefaults()
}

func (cmd *initCmd) parseAndExec(args []string) error {
	done, err := parseStartHelp(&cmd.flags, &cmd.option, cmd.err, args, true)
	if done || err != nil {
		return err
	}

	var print bool
	for _, arg := range cmd.flags.Args() {
		switch arg {
		case "-":
			print = true
		default:
			cmd.shell = arg
			cmd.resetShell()
		}
	}

	if print {
		cmd.printInitShellScripts(cmd.out)
	} else {
		cmd.flags.Usage()
		return nil
	}

	return nil
}

func (cmd *initCmd) resetShell() {
	cmd.shProf = shellProfile(cmd.shell)
	cmd.flags.Usage = cmd.usage
}

func (cmd *initCmd) printInitShellScripts(out io.Writer) {
	fmt.Fprintf(
		out, `export %s="%s"
PATH="%s:${PATH}"

# Load script in a package
include() {
  _package="$1"
  _file="$2"

  if [ -z "${_package}" ] || [ -z "${_file}" ]; then
    echo "Usage: include <package> <file>" >&2
    unset _package _file
    return 1
  fi

  if [ ! -e "${%s}/packages/${_package}" ]; then
    echo "Package not installed: ${_package}" >&2
    unset _package _file
    return 1
  fi

  if [ -e "${%s}/packages/${_package}/${_file}" ]; then
    . "${%s}/packages/${_package}/${_file}" >&2
    unset _package _file
  else
    echo "File not found: ${%s}/packages/${_package}/${_file}" >&2
    unset _package _file
    return 1
  fi
}
`,
		config.RootVarName, cmd.config.RootPath(), cmd.config.BinPath(),
		config.RootVarName, config.RootVarName, config.RootVarName, config.RootVarName)
}
