package cli

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"golang.org/x/crypto/ssh/terminal"
)

type pruneCmd struct {
	commonCmd
	option struct {
		verboseFlags
		yes  *bool
		link *bool
	}
}

func (cmd *pruneCmd) getOpts() flagger {
	return &cmd.option
}

func (cmd *pruneCmd) verboseOpts() verboseFlagger {
	return &cmd.option
}

const (
	defined = iota + 1
	prunee  // to be pruned
)

func newPruneCmd(common commonCmd) pruneCmd {
	cmd := &pruneCmd{}
	cmd.commonCmd = common
	setupCmdFlags(cmd, "prune", cmd.usage)
	cmd.option.yes = cmd.flags.BoolP("yes", "y", false, "# Prune without confirmation")
	cmd.option.link = cmd.flags.Bool("link", false, "# Prune symlinks at the same time")
	return *cmd
}

func (cmd *pruneCmd) usage() {
	const help = `Summary:
  Uninstall packages not defined in config file.

Syntax:
  {{.Prog}} {{.Cmd}} [-c|--config CONFIG] [-y|--yes] [--link]

This doesn't remove symlinks created with "link" command by default.
To remove them, specify "--link" option.

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "prune"})
	cmd.flags.PrintDefaults()
}

func (cmd *pruneCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
	if done || err != nil {
		return err
	}

	founds, err := installedPackages(cmd, true)
	if err != nil {
		return err
	}
	candidates := make(map[string]int)

	skipped := 0
	for _, fi := range founds {
		path := filepath.Join(cmd.config.PackagePath(), fi.Name())
		if *cmd.option.link {
			candidates[fi.Name()] = prunee
			continue
		}

		link, err := os.Readlink(path)
		if link != "" {
			if err != nil {
				// Just in case
				fmt.Fprintf(cmd.err, "Error! Reading link failed. Path = %s\n", path)
			}
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.err, "\"%s\" is symlink. Skip\n", fi.Name())
			}
			skipped++
			continue
		}
		candidates[fi.Name()] = prunee
	}

	definedCnt := 0
	for _, param := range cmd.config.Packages {
		if param.From == "" {
			fmt.Fprintf(cmd.err, "Warning! \"from\" is not specified. Skips. pkg = %+v\n", param)
		}

		pkg, err := packageToInstall(cmd, installArgs{param.From, param.As, param.At, param.Bin})
		if err == nil && candidates[pkg.name] == prunee {
			if *cmd.option.verbose {
				fmt.Fprintf(cmd.err, "\"%s\" is configured\n", pkg.name)
			}
			candidates[pkg.name] = defined
			definedCnt++
		}
	}

	if len(founds)-skipped == definedCnt {
		// All existing packages are configured
		return ErrCanceled
	}

	prunees := []string{}
	for name, val := range candidates {
		if val == defined {
			continue
		}
		prunees = append(prunees, name)
	}

	if terminal.IsTerminal(0) && !*cmd.option.yes {
		const confirmation = `Packages to remove:
{{- range $i, $name := .Packages}}
  {{$name}}{{end}}

Okay? (Y/n) `
		t := template.Must(template.New("usage").Parse(confirmation))
		t.Execute(cmd.out, struct{ Packages []string }{prunees})
		stdin := bufio.NewScanner(os.Stdin)
		stdin.Scan()
		input := stdin.Text()
		if strings.HasPrefix(input, "n") || strings.HasPrefix(input, "N") {
			fmt.Fprintln(cmd.out, "Canceled")
			return ErrCanceled
		}
	}

	for _, name := range prunees {
		if err = removePackage(cmd, name); err != nil {
			return err
		}
	}

	return nil
}
