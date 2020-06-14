package cli

import (
	"fmt"
	"text/template"

	"github.com/progrhyme/shelp/internal/git"
)

type bundleCmd struct {
	gitCmd
}

func newBundleCmd(common commonCmd, git git.Git) bundleCmd {
	cmd := &bundleCmd{}
	cmd.commonCmd = common
	cmd.git = git
	setupCmdFlags(cmd, "bundle", cmd.usage)
	return *cmd
}

func (cmd *bundleCmd) usage() {
	const help = `Summary:
  Install packages at once which are defined in config file.

Syntax:
  {{.Prog}} {{.Cmd}}

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.err, struct{ Prog, Cmd string }{cmd.command, "bundle"})
	cmd.flags.PrintDefaults()
	fmt.Fprintf(cmd.err, `
Limitation:
  Re-installation of existing package is not supported yet.
  To do this, you have to remove it beforehand.
`)
}

func (cmd *bundleCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false)
	if done || err != nil {
		return err
	}

	if err = prepareInstallDirectories(cmd.config); err != nil {
		fmt.Fprintf(cmd.err, "Error! %s\n", err)
		return ErrOperationFailed
	}

	var (
		success  int
		hasError bool
	)
	for _, param := range cmd.config.Packages {
		if param.From != "" {
			// Install package by Git
			err = installPackage(cmd, installArgs{param.From, param.As, param.At, param.Bin})
			switch err {
			case nil, ErrAlreadyInstalled:
				success++
			default:
				hasError = true
			}
			//} else if param.Link != "" {
			// Link local package
			//fmt.Fprintf(cmd.err, "TODO: link is not implemented yet. pkg = %+v\n", param)
		} else {
			//fmt.Fprintf(cmd.err, "Warning! Neither \"from\" nor \"link\" is specified. pkg = %+v\n", param)
			fmt.Fprintf(cmd.err, "Warning! \"from\" is not specified. Skips. pkg = %+v\n", param)
			hasError = true
		}
	}

	if hasError {
		if success > 0 {
			fmt.Fprintln(cmd.err, "There are some errors")
			return ErrWarning
		} else {
			fmt.Fprintln(cmd.err, "Bundle failed")
			return ErrOperationFailed
		}
	}

	return nil
}