package main

import (
	"os"

	"github.com/progrhyme/claft/internal/cli"
	"github.com/progrhyme/claft/internal/config"
	"github.com/progrhyme/claft/internal/git"
)

func main() {
	cfg := config.NewConfig()
	g := git.NewGit(os.Stdout, os.Stderr)
	c := cli.NewCli(version, cfg, g, os.Stdout, os.Stderr)
	e := c.ParseAndExec(os.Args)

	if e != nil {
		os.Exit(1)
	}
}
