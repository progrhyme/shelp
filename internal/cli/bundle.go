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
  {{.Prog}} {{.Cmd}} [-c|--config CONFIG]

Options:
`

	t := template.Must(template.New("usage").Parse(help))
	t.Execute(cmd.errs, struct{ Prog, Cmd string }{cmd.name, "bundle"})
	cmd.flags.PrintDefaults()
	fmt.Fprintf(cmd.errs, `
Specification:
  If there are pseudo-installed packages created by "link" command among configured packages,
  "bundle" operation re-install the packages.
`)
}

func (cmd *bundleCmd) parseAndExec(args []string) error {
	done, err := parseStart(cmd, args, false, false)
	if done || err != nil {
		return err
	}

	if len(cmd.config.Packages) == 0 {
		fmt.Fprintln(cmd.errs, "No package is configured")
		cmd.flags.Usage()
		return ErrCanceled
	}

	if err = prepareInstallDirectories(cmd.config); err != nil {
		fmt.Fprintf(cmd.errs, "Error! %s\n", err)
		return ErrOperationFailed
	}

	var (
		success  int
		hasError bool
	)
	for _, param := range cmd.config.Packages {
		if param.From == "" {
			fmt.Fprintf(cmd.errs, "Warning! \"from\" is not specified. Skips. pkg = %+v\n", param)
			hasError = true
		}

		// Install one
		pkg := installArgs{
			from:      param.From,
			as:        param.As,
			at:        param.At,
			bin:       param.Bin,
			overwrite: true,
		}
		err = installPackage(cmd, pkg)
		switch err {
		case nil, ErrAlreadyInstalled:
			success++
		default:
			hasError = true
		}
	}

	if hasError {
		if success > 0 {
			fmt.Fprintln(cmd.errs, "There are some errors")
			return ErrWarning
		} else {
			fmt.Fprintln(cmd.errs, "Bundle failed")
			return ErrOperationFailed
		}
	}

	return nil
}
