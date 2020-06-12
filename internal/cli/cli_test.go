package cli

import (
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
)

// Tests for all available commands using table-driven tests
// but no operation which affects filesystem
func TestParseAndExecAll(t *testing.T) {
	// Setup
	prog := "shelp"
	version := "0.0.1"
	os.Setenv(config.RootVarName, "tmp")
	os.MkdirAll("tmp", 0755)
	cfg := config.NewConfig()

	// Variables as test parameters or expected outputs
	validPkgRepo := "bats-core/bats-core"
	invalidFlg := "--no-such-option"
	flagError := fmt.Sprintf("Error! unknown flag: %s", invalidFlg)

	type command struct {
		requireArg bool
		helpText   string
	}
	commands := make(map[string]command)

	commands["root"] = command{
		false,
		fmt.Sprintf(`Summary:
  "%s" is a Git-based package manager for shell scripts written in Go.

Usage:`, prog),
	}

	commands["init"] = command{
		true,
		fmt.Sprintf(`Summary:
  Enable %s in one's shell environment.

Usage:`, prog),
	}

	commands["install"] = command{
		true,
		fmt.Sprintf(`Summary:
  Install a repository from HTTPS site as a %s package, assuming it contains shell scripts.

Syntax:`, prog),
	}

	commands["remove"] = command{
		true,
		`Summary:
  Uninstall a package.

Syntax:`,
	}

	commands["list"] = command{
		false,
		`Summary:
  List installed packages.

Syntax:`,
	}

	commands["upgrade"] = command{
		true,
		`Summary:
  Upgrade an installed package.

Syntax:`,
	}

	commands["outdated"] = command{
		true,
		`Summary:
  Show installed packages which can be updated.

Syntax:`,
	}

	commands["link"] = command{
		true,
		`Summary:
  Pseudo installation of a package from local filesystem.
  Creates symbolic link of a directory into a package path.

Syntax:`,
	}

	commands["destroy"] = command{
		false,
		fmt.Sprintf(`Summary:
  Delete all contents in %s including the root directory.

Syntax:`, config.RootVarName),
	}

	initText := fmt.Sprintf(`export %s="%s"
PATH="%s:${PATH}"

# Load script in a package`, config.RootVarName, cfg.RootPath(), cfg.BinPath())

	// Test cases
	tests := []struct {
		args   []string
		err    error
		outStr string
		errStr string
	}{
		// Without subcommand
		{[]string{prog}, ErrUsage, "", commands["root"].helpText},
		{
			[]string{prog, "--help"},
			nil, "", commands["root"].helpText,
		},
		{
			[]string{prog, "--version"},
			nil, version, "",
		},
		{
			[]string{prog, "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["root"].helpText}, "\n"),
		},

		// Subcommand "init"
		{
			[]string{prog, "init"},
			ErrUsage, "", commands["init"].helpText,
		},
		{
			[]string{prog, "init", "--help"},
			nil, "", commands["init"].helpText,
		},
		{
			[]string{prog, "init", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["init"].helpText}, "\n"),
		},
		{[]string{prog, "init", "-"}, nil, initText, ""},

		// Subcommand "install"
		{
			[]string{prog, "install"},
			ErrUsage, "", commands["install"].helpText,
		},
		{
			[]string{prog, "install", "--help"},
			nil, "", commands["install"].helpText,
		},
		{
			[]string{prog, "install", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["install"].helpText}, "\n"),
		},
		{
			[]string{prog, "install", "invalid-repo-specifier"},
			ErrArgument, "",
			strings.Join([]string{"Error! Given argument \"invalid-repo-specifier\" does not look like valid repository", commands["install"].helpText}, "\n"),
		},
		{
			[]string{prog, "install", validPkgRepo, "-"},
			ErrArgument, "",
			"Error! Given argument \"-\" does not look like valid package name\n",
		},

		// Subcommand "remove"
		{
			[]string{prog, "remove"},
			ErrUsage, "", commands["remove"].helpText,
		},
		{
			[]string{prog, "remove", "--help"},
			nil, "", commands["remove"].helpText,
		},
		{
			[]string{prog, "remove", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["remove"].helpText}, "\n"),
		},
		{
			[]string{prog, "remove", "not-installed-package"},
			ErrArgument, "",
			"\"not-installed-package\" is not installed",
		},

		// Subcommand "upgrade"
		{
			[]string{prog, "upgrade"},
			ErrUsage, "", commands["upgrade"].helpText,
		},
		{
			[]string{prog, "upgrade", "--help"},
			nil, "", commands["upgrade"].helpText,
		},
		{
			[]string{prog, "upgrade", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["upgrade"].helpText}, "\n"),
		},
		{
			[]string{prog, "upgrade", "not-installed-package"},
			ErrArgument, "",
			"\"not-installed-package\" is not installed",
		},

		// Subcommand "outdated"
		{
			[]string{prog, "outdated"},
			nil, "", "No package is installed",
		},
		{
			[]string{prog, "outdated", "--help"},
			nil, "", commands["outdated"].helpText,
		},
		{
			[]string{prog, "outdated", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["outdated"].helpText}, "\n"),
		},

		// Subcommand "list"
		{
			[]string{prog, "list"},
			nil, "", "No package is installed",
		},
		{
			[]string{prog, "list", "--help"},
			nil, "", commands["list"].helpText,
		},
		{
			[]string{prog, "list", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["list"].helpText}, "\n"),
		},

		// Subcommand "link"
		{
			[]string{prog, "link"},
			ErrUsage, "", commands["link"].helpText,
		},
		{
			[]string{prog, "link", "--help"},
			nil, "", commands["link"].helpText,
		},
		{
			[]string{prog, "link", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["link"].helpText}, "\n"),
		},
		{
			[]string{prog, "link", "no/such/file/or/directory"},
			ErrArgument, "",
			"Error! \"no/such/file/or/directory\" does not exist\n",
		},
		{
			[]string{prog, "link", ".", "-"},
			ErrArgument, "",
			"Error! Given argument \"-\" does not look like valid package name\n",
		},

		// Subcommand "destroy"
		{
			[]string{prog, "destroy"},
			ErrOperationFailed, "", "Warning! Destruction is canceled because flag \"yes\" is not set",
		},
		{
			[]string{prog, "destroy", "--help"},
			nil, "", commands["destroy"].helpText,
		},
		{
			[]string{prog, "destroy", "--no-such-option"},
			ErrParseFailed, "",
			strings.Join([]string{flagError, commands["destroy"].helpText}, "\n"),
		},
	}

	// Run test cases
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			out := &strings.Builder{}
			err := &strings.Builder{}
			g := git.NewGit(out, err)
			c := NewCli(version, cfg, g, out, err)
			e := c.ParseAndExec(tt.args)
			if e != tt.err {
				t.Errorf("[Ret] Got: %v, Expected: %v", e, tt.err)
			}
			if tt.outStr == "" {
				if out.String() != "" {
					t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tt.outStr)
				}
			} else if !strings.Contains(out.String(), tt.outStr) {
				t.Errorf("[Stdout] Got: %s, Expected: %s", out.String(), tt.outStr)
			}
			if tt.errStr == "" {
				if err.String() != "" {
					t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tt.errStr)
				}
			} else if !strings.Contains(err.String(), tt.errStr) {
				t.Errorf("[Stderr] Got: %s, Expected: %s", err.String(), tt.errStr)
			}
		})
	}
}
