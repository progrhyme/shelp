package cli

import (
	"io"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

type commander interface {
	getConf() *config.Config
	flagset() *pflag.FlagSet
	outs() io.Writer
	errs() io.Writer
}

// Meets commander interface
type commonCmd struct {
	config  config.Config
	flags   pflag.FlagSet
	out     io.Writer
	err     io.Writer
	command string
}

func (cmd *commonCmd) getConf() *config.Config {
	return &cmd.config
}

func (cmd *commonCmd) flagset() *pflag.FlagSet {
	return &cmd.flags
}

func (cmd *commonCmd) outs() io.Writer {
	return cmd.out
}

func (cmd *commonCmd) errs() io.Writer {
	return cmd.err
}

type flagger interface {
	helpFlg() *bool
}

type commonFlags struct {
	help *bool
}

func (flag *commonFlags) helpFlg() *bool {
	return flag.help
}

type helpCommander interface {
	commander
	getOpts() flagger
}

type helpCmd struct {
	commonCmd
	option commonFlags
}

func (cmd *helpCmd) getOpts() flagger {
	return &cmd.option
}

type verboseFlagger interface {
	flagger
	verboseFlg() *bool
}

type verboseFlags struct {
	commonFlags
	verbose *bool
}

func (flag *verboseFlags) verboseFlg() *bool {
	return flag.verbose
}

type verboseCommander interface {
	commander
	verboseOpts() verboseFlagger
}

type verboseCmd struct {
	commonCmd
	option verboseFlags
}

func (cmd *verboseCmd) getOpts() flagger {
	return &cmd.option
}

func (cmd *verboseCmd) verboseOpts() verboseFlagger {
	return &cmd.option
}

type gitCommander interface {
	verboseCommander
	gitCtl() *git.Git
}

type gitCmd struct {
	verboseCmd
	git git.Git
}

func (cmd *gitCmd) gitCtl() *git.Git {
	return &cmd.git
}
