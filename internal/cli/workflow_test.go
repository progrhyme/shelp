package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/progrhyme/shelp/internal/config"
	"github.com/progrhyme/shelp/internal/git"
)

type cliParam struct {
	prog    string
	version string
	config  config.Config
}

type installParam struct {
	args []string
	pkg  string
	bins []string
}

type removeParam struct {
	pkg  string
	bins []string
}

type linkParam struct {
	args []string
	pkg  string
	bins []string
}

// Run typical workflow of `shelp` and validate outputs and state of filesystem including installed
// packages
func TestWorkflow(t *testing.T) {
	// Setup
	prog := "shelp"
	version := "0.0.1"
	tmpDir, err := filepath.Abs("tmp")
	if err != nil {
		panic(fmt.Sprintf("Can't resolve \"tmp\"! Error = %v", err))
	}
	rootDir := filepath.Join(tmpDir, "shelp")
	srcDir := filepath.Join(tmpDir, "src")
	os.Setenv(config.RootVarName, rootDir)
	os.MkdirAll(rootDir, 0755)
	os.MkdirAll(srcDir, 0755)
	cfg := config.NewConfig()

	cparam := cliParam{prog, version, cfg}

	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Can't get pwd! Error = %v", err))
	}
	repoDir := strings.TrimSuffix(pwd, fmt.Sprintf("%cinternal%ccli", os.PathSeparator, os.PathSeparator))

	// Install some packages
	installParams := []installParam{
		{
			args: []string{"progrhyme/bash-links"},
			pkg:  "bash-links",
			bins: []string{"links"},
		},
		{
			args: []string{"github.com/progrhyme/shove@v0.8.3", "shove@v0.8.3"},
			pkg:  "shove@v0.8.3",
			bins: []string{"shove"},
		},
		{
			args: []string{fmt.Sprintf("file://%s", repoDir)},
			pkg:  "shelp",
			bins: []string{},
		},
	}
	testInstallPackages(t, installParams, cparam)

	testListInstalled(t, []string{"bash-links", "shelp", "shove@v0.8.3"}, cparam)

	// prepare source contents before remove
	outStr := &strings.Builder{}
	errStr := &strings.Builder{}
	gitCtl := git.NewGit(outStr, errStr)
	installedPkgURL := fmt.Sprintf("file://%s", filepath.Join(cfg.PackagePath(), "bash-links"))
	linkSrc := filepath.Join(srcDir, "bash-links")
	err = gitCtl.Clone(installedPkgURL, linkSrc, "", false)
	if err != nil {
		panic(fmt.Sprintf("Failed to git clone %s to %s", installedPkgURL, linkSrc))
	}

	// Remove some packages
	removeParams := []removeParam{
		{pkg: "bash-links", bins: []string{"links"}},
	}
	testRemovePackages(t, removeParams, cparam)

	testListInstalled(t, []string{"shelp", "shove@v0.8.3"}, cparam)

	// Link some packages
	linkParams := []linkParam{
		{
			args: []string{"."},
			pkg:  "cli",
			bins: []string{},
		},
		{
			args: []string{linkSrc, "links"},
			pkg:  "links",
			bins: []string{"links"},
		},
	}
	testLinkPackages(t, linkParams, cparam)

	testListInstalled(t, []string{"cli", "links", "shelp", "shove@v0.8.3"}, cparam)

	removeParams = []removeParam{
		{pkg: "links", bins: []string{"links"}},
	}
	testRemovePackages(t, removeParams, cparam)

	testListInstalled(t, []string{"cli", "shelp", "shove@v0.8.3"}, cparam)

	testDestroy(t, cparam)
	testListInstalled(t, []string{}, cparam)

	// Clean up
	os.RemoveAll(tmpDir)
}

func testInstallPackages(t *testing.T, targets []installParam, cp cliParam) {
	for _, target := range targets {
		subtest := fmt.Sprintf("install %s", strings.Join(target.args, " "))
		t.Run(subtest, func(t *testing.T) {
			outStr := &strings.Builder{}
			errStr := &strings.Builder{}
			gitCtl := git.NewGit(outStr, errStr)
			ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)
			args := append([]string{cp.prog, "install"}, target.args...)

			err := ctl.ParseAndExec(args)
			if err != nil {
				t.Errorf("Install failed. Error = %v, args = %v", err, target.args)
			}
			path := filepath.Join(cp.config.PackagePath(), target.pkg)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("\"%s\" does not exist", path)
			}
			for _, bin := range target.bins {
				sym := filepath.Join(cp.config.BinPath(), bin)
				src, err := os.Readlink(sym)
				if err != nil {
					t.Errorf("Bin not installed: %s. Error = %s", bin, err)
					continue
				}
				if !strings.HasPrefix(src, path) {
					t.Errorf("Not linked to package: %s -> %s", sym, src)
				}
			}
		})
	}
}

func testLinkPackages(t *testing.T, targets []linkParam, cp cliParam) {
	for _, target := range targets {
		subtest := fmt.Sprintf("link %s", strings.Join(target.args, " "))
		t.Run(subtest, func(t *testing.T) {
			outStr := &strings.Builder{}
			errStr := &strings.Builder{}
			gitCtl := git.NewGit(outStr, errStr)
			ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)
			args := append([]string{cp.prog, "link"}, target.args...)

			err := ctl.ParseAndExec(args)
			if err != nil {
				t.Errorf("Link failed. Error = %v, args = %v", err, target.args)
			}
			path := filepath.Join(cp.config.PackagePath(), target.pkg)
			if _, err := os.Stat(path); os.IsNotExist(err) {
				t.Errorf("\"%s\" does not exist", path)
			}
			sym := filepath.Join(cp.config.PackagePath(), target.pkg)
			src, err := os.Readlink(sym)
			if err != nil {
				t.Errorf("Package not linked: %s. Error = %s", target.pkg, err)
				return
			}
			abs, err := filepath.Abs(target.args[0])
			if err != nil {
				panic(fmt.Sprintf("Failed to resolve path of %s. Error = %v", target.args[0], err))
			}
			if !strings.HasPrefix(src, abs) {
				t.Errorf("Not linked to src: %s -> %s", sym, target.args[0])
			}
			for _, bin := range target.bins {
				sym := filepath.Join(cp.config.BinPath(), bin)
				src, err := os.Readlink(sym)
				if err != nil {
					t.Errorf("Bin not linked: %s. Error = %s", bin, err)
					continue
				}
				if !strings.HasPrefix(src, path) {
					t.Errorf("Not linked to package: %s -> %s", sym, src)
				}
			}
		})
	}
}

func testListInstalled(t *testing.T, pkgs []string, cp cliParam) {
	t.Run("list", func(t *testing.T) {
		outStr := &strings.Builder{}
		errStr := &strings.Builder{}
		gitCtl := git.NewGit(outStr, errStr)
		ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)

		err := ctl.ParseAndExec([]string{cp.prog, "list"})
		if err != nil {
			t.Errorf("List failed. Error = %v", err)
		}
		expected := ""
		for _, pkg := range pkgs {
			expected += fmt.Sprintf("%s\n", pkg)
		}
		if outStr.String() != expected {
			t.Errorf("[Stdout] Got: %s, Expected: %s", outStr.String(), expected)
		}
	})
}

func testRemovePackages(t *testing.T, targets []removeParam, cp cliParam) {
	for _, target := range targets {
		subtest := fmt.Sprintf("remove %s", target.pkg)
		t.Run(subtest, func(t *testing.T) {
			outStr := &strings.Builder{}
			errStr := &strings.Builder{}
			gitCtl := git.NewGit(outStr, errStr)
			ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)
			args := append([]string{cp.prog, "remove"}, target.pkg)

			err := ctl.ParseAndExec(args)
			if err != nil {
				t.Errorf("Remove failed. Error = %v, target = %s", err, target.pkg)
			}
			path := filepath.Join(cp.config.PackagePath(), target.pkg)
			if _, err := os.Stat(path); !os.IsNotExist(err) {
				t.Errorf("\"%s\" stil exists", path)
			}
			for _, bin := range target.bins {
				sym := filepath.Join(cp.config.BinPath(), bin)
				if _, err := os.Stat(sym); !os.IsNotExist(err) {
					t.Errorf("\"%s\" stil exists", sym)
				}
			}
		})
	}
}

func testDestroy(t *testing.T, cp cliParam) {
	args := []string{"destroy", "--yes"}
	t.Run(strings.Join(args, " "), func(t *testing.T) {
		outStr := &strings.Builder{}
		errStr := &strings.Builder{}
		gitCtl := git.NewGit(outStr, errStr)
		ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)
		args = append([]string{cp.prog}, args...)

		err := ctl.ParseAndExec(args)
		if err != nil {
			t.Errorf("Destroy failed. Error = %v", err)
		}
		if _, err := os.Stat(cp.config.RootPath()); !os.IsNotExist(err) {
			t.Errorf("\"%s\" stil exists", cp.config.RootPath())
		}
	})
}
