package main

import (
	"os"

	"github.com/progrhyme/shelp/internal/cli"
	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
)

func main() {
	cfg := config.NewConfig()
	g := git.NewGit(os.Stdout, os.Stderr)
	c := cli.NewCli(version, cfg, g, os.Stdout, os.Stderr)
	e := c.ParseAndExec(os.Args)

	switch e {
	case nil, cli.ErrCanceled:
		// OK
	default:
		os.Exit(1)
	}
}
