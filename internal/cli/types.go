package cli

import (
	"io"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
	"github.com/spf13/pflag"
)

type runner interface {
	getConfig() *config.Config
	getOuts() io.Writer
	getErrs() io.Writer
	getFlags() *pflag.FlagSet
	setFlags(*pflag.FlagSet)
}

// Meets runner interface
type commonCmd struct {
	config *config.Config
	flags  pflag.FlagSet
	outs   io.Writer
	errs   io.Writer
	name   string
}

func (cmd *commonCmd) getConfig() *config.Config {
	return cmd.config
}

func (cmd *commonCmd) getOuts() io.Writer {
	return cmd.outs
}

func (cmd *commonCmd) getErrs() io.Writer {
	return cmd.errs
}

func (cmd *commonCmd) getFlags() *pflag.FlagSet {
	return &cmd.flags
}

func (cmd *commonCmd) setFlags(flags *pflag.FlagSet) {
	cmd.flags = *flags
}

type flavor interface {
	getHelp() *bool
	getConfig() *string
	setHelp(*bool)
	setConfig(*string)
}

type commonOpts struct {
	help   *bool
	config *string
}

func (flag *commonOpts) getHelp() *bool {
	return flag.help
}

func (flag *commonOpts) getConfig() *string {
	return flag.config
}

func (flag *commonOpts) setHelp(help *bool) {
	flag.help = help
}

func (flag *commonOpts) setConfig(conf *string) {
	flag.config = conf
}

type helpRunner interface {
	runner
	getOpts() flavor
}

type helpCmd struct {
	commonCmd
	option commonOpts
}

func (cmd *helpCmd) getOpts() flavor {
	return &cmd.option
}

type verboseFlavor interface {
	flavor
	getVerbose() *bool
	setVerbose(*bool)
}

type verboseOpts struct {
	commonOpts
	verbose *bool
}

func (flag *verboseOpts) getVerbose() *bool {
	return flag.verbose
}

func (flag *verboseOpts) setVerbose(verbose *bool) {
	flag.verbose = verbose
}

type verboseRunner interface {
	runner
	getVerboseOpts() verboseFlavor
}

// verboseCmd meets both helpRunner & verboseRunner interfaces
type verboseCmd struct {
	commonCmd
	option verboseOpts
}

func (cmd *verboseCmd) getOpts() flavor {
	return &cmd.option
}

func (cmd *verboseCmd) getVerboseOpts() verboseFlavor {
	return &cmd.option
}

type gitRunner interface {
	verboseRunner
	getGit() *git.Git
}

type gitCmd struct {
	verboseCmd
	git git.Git
}

func (cmd *gitCmd) getGit() *git.Git {
	return &cmd.git
}
