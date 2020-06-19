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

type testCliParam struct {
	prog    string
	version string
	config  *config.Config
}

type testInstallParam struct {
	args []string
	pkg  string
	bins []string
}

type testUpgradeParam struct {
	pkg    string
	outStr string
}

type testRemoveParam struct {
	pkg  string
	bins []string
}

type testLinkParam struct {
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
	cfg := config.NewConfig(os.Stdout, os.Stderr)

	cparam := testCliParam{prog, version, &cfg}

	pwd, err := os.Getwd()
	if err != nil {
		panic(fmt.Sprintf("Can't get pwd! Error = %v", err))
	}
	repoDir := strings.TrimSuffix(pwd, fmt.Sprintf("%cinternal%ccli", os.PathSeparator, os.PathSeparator))

	// Install some packages
	testInstallParams := []testInstallParam{
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
			args: []string{"github.com/progrhyme/gcloud-prompt@1d2f918"},
			pkg:  "gcloud-prompt",
			bins: []string{},
		},
		{
			args: []string{fmt.Sprintf("file://%s", repoDir)},
			pkg:  "shelp",
			bins: []string{},
		},
	}
	testInstallPackages(t, testInstallParams, cparam)

	testListInstalled(t, []string{"bash-links", "gcloud-prompt", "shelp", "shove@v0.8.3"}, cparam)

	// Upgrade packages
	testUpgradeParams := []testUpgradeParam{
		{pkg: "bash-links", outStr: "No need to upgrade"},
		{pkg: "gcloud-prompt", outStr: "No need to upgrade"},
		{pkg: "shove@v0.8.3", outStr: "No need to upgrade"},
	}
	testUpgradePackages(t, testUpgradeParams, cparam)

	// prepare source contents before remove
	outStr := &strings.Builder{}
	errStr := &strings.Builder{}
	gitCtl := git.NewGit(outStr, errStr)
	installedPkgURL := fmt.Sprintf("file://%s", filepath.Join(cfg.PackagePath(), "bash-links"))
	linkSrc := filepath.Join(srcDir, "bash-links")
	err = gitCtl.Clone(installedPkgURL, linkSrc, git.Option{})
	if err != nil {
		panic(fmt.Sprintf("Failed to git clone %s to %s", installedPkgURL, linkSrc))
	}

	// Remove some packages
	testRemoveParams := []testRemoveParam{
		{pkg: "bash-links", bins: []string{"links"}},
	}
	testRemovePackages(t, testRemoveParams, cparam)

	testListInstalled(t, []string{"gcloud-prompt", "shelp", "shove@v0.8.3"}, cparam)

	// Link some packages
	testLinkParams := []testLinkParam{
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
	testLinkPackages(t, testLinkParams, cparam)

	testListInstalled(t, []string{"cli", "gcloud-prompt", "links", "shelp", "shove@v0.8.3"}, cparam)

	testRemoveParams = []testRemoveParam{
		{pkg: "links", bins: []string{"links"}},
	}
	testRemovePackages(t, testRemoveParams, cparam)

	testListInstalled(t, []string{"cli", "gcloud-prompt", "shelp", "shove@v0.8.3"}, cparam)

	testDestroy(t, cparam)
	testListInstalled(t, []string{}, cparam)

	// Clean up
	os.RemoveAll(tmpDir)
}

func testInstallPackages(t *testing.T, targets []testInstallParam, cp testCliParam) {
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

func testUpgradePackages(t *testing.T, targets []testUpgradeParam, cp testCliParam) {
	for _, target := range targets {
		subtest := fmt.Sprintf("upgrade %s", target.pkg)
		t.Run(subtest, func(t *testing.T) {
			outStr := &strings.Builder{}
			errStr := &strings.Builder{}
			gitCtl := git.NewGit(outStr, errStr)
			ctl := NewCli(cp.version, cp.config, gitCtl, outStr, errStr)
			args := append([]string{cp.prog, "upgrade"}, target.pkg)

			err := ctl.ParseAndExec(args)
			if err != nil {
				t.Errorf("Upgrade failed. Error = %v, target = %s", err, target.pkg)
			}
			if !strings.Contains(outStr.String(), target.outStr) {
				t.Errorf("[Stdout] Got: %s, Expected: %s", outStr.String(), target.outStr)
			}
		})
	}
}

func testLinkPackages(t *testing.T, targets []testLinkParam, cp testCliParam) {
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

func testListInstalled(t *testing.T, pkgs []string, cp testCliParam) {
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

func testRemovePackages(t *testing.T, targets []testRemoveParam, cp testCliParam) {
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

func testDestroy(t *testing.T, cp testCliParam) {
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
