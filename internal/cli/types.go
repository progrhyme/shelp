package cli

import (
	"io"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/spf13/pflag"
)

type flagger interface {
	helpFlg() *bool
}

type commander interface {
	getConf() *config.Config
	outs() io.Writer
	errs() io.Writer
}

type verboseFlagger interface {
	flagger
	verboseFlg() *bool
}

type verboseCommander interface {
	commander
	getOpts() verboseFlagger
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

func (cmd *commonCmd) outs() io.Writer {
	return cmd.out
}

func (cmd *commonCmd) errs() io.Writer {
	return cmd.err
}

type commonFlags struct {
	help *bool
}

func (flag *commonFlags) helpFlg() *bool {
	return flag.help
}

type verboseFlags struct {
	commonFlags
	verbose *bool
}

func (flag *verboseFlags) verboseFlg() *bool {
	return flag.verbose
}
