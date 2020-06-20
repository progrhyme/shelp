package cli

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"text/template"

	"github.com/progrhyme/shelp/internal/config"
)

type initCmd struct {
	helpCmd
	shell  string
	shProf string
}

func newInitCmd(common commonCmd) initCmd {
	cmd := &initCmd{}
	cmd.commonCmd = common

	cmd.shell = filepath.Base(os.Getenv("SHELL"))
	cmd.shProf = shellProfile(cmd.shell)

	setupCmdFlags(cmd, "init", cmd.usage)
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
	case "fish":
		prof = "~/.config/fish/config.fish"
	default:
		prof = "its profile"
	}
	return prof
}

func (cmd *initCmd) usage() {
	const help = `Summary:
  Enable {{.Prog}} in one's shell environment.

Usage:
    {{.Prog}} init - [SHELL]  # Print scripts (for specified SHELL)

  To enable {{.Prog}} automatically in one's shell, append the following to {{.Profile}}:

    {{.InitCommand}}

  It prints scripts for current shell unless user specify SHELL argument.

Supported Shells:
- Most POSIX compatible shells including Zsh
- fish shell

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	params := struct{ Prog, Profile, InitCommand string }{
		Prog: cmd.name, Profile: cmd.shProf,
	}
	switch cmd.shell {
	case "fish":
		params.InitCommand = fmt.Sprintf(`%s init - | source`, cmd.name)
	default:
		params.InitCommand = fmt.Sprintf(`eval "$(%s init -)`, cmd.name)
	}
	t.Execute(cmd.errs, params)

	cmd.flags.PrintDefaults()
}

func (cmd *initCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, true, true)
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
		cmd.printInitShellScripts(cmd.outs)
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
	var script string
	switch cmd.shell {
	case "fish":
		script = `set -gx <<.RootPathKey>> <<.RootPath>>
if not contains <<.BinPath>> $PATH
  set -gx PATH <<.BinPath>> $PATH
end

# Load script in a package
function include
  set package $argv[1]
  set file $argv[2]

  if test -z "$package" -o -z "$file"
    echo "Usage: include <package> <file>" >&2
    return 1
  end

  if test ! -e "$<<.RootPathKey>>/packages/$package"
    echo "Package not installed: $package" >&2
    return 1
  end

  if test -e "$<<.RootPathKey>>/packages/$package/$file"
    source "$<<.RootPathKey>>/packages/$package/$file" >&2
  else
    echo "File not found: $<<.RootPathKey>>/packages/$package/$file" >&2
    return 1
  end
end
`

	default:
		script = `export <<.RootPathKey>>="<<.RootPath>>"
PATH="<<.BinPath>>:${PATH}"

# Load script in a package
include() {
  _package="$1"
  _file="$2"

  if [ -z "${_package}" ] || [ -z "${_file}" ]; then
    echo "Usage: include <package> <file>" >&2
    unset _package _file
    return 1
  fi

  if [ ! -e "${<<.RootPathKey>>}/packages/${_package}" ]; then
    echo "Package not installed: ${_package}" >&2
    unset _package _file
    return 1
  fi

  if [ -e "${<<.RootPathKey>>}/packages/${_package}/${_file}" ]; then
    . "${<<.RootPathKey>>}/packages/${_package}/${_file}" >&2
    unset _package _file
  else
    echo "File not found: ${<<.RootPathKey>>}/packages/${_package}/${_file}" >&2
    unset _package _file
    return 1
  fi
}
`
	}

	params := struct{ RootPathKey, RootPath, BinPath string }{
		config.RootVarName, cmd.config.RootPath(), cmd.config.BinPath()}
	t := template.Must(template.New("script").Delims("<<", ">>").Parse(script))
	t.Execute(out, params)
}
