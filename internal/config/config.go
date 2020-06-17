package config

import (
	"fmt"
	"io"
	"os"
	"path/filepath"

	"gopkg.in/yaml.v2"
)

// Configuration properties set by config file
type properties struct {
	Path struct {
		Root string
	}
	Git struct {
		Shallow bool
	}
	Packages []struct {
		From string
		Bin  []string
		As   string
		At   string
	}
}

// Config wraps embedded properties
type Config struct {
	outs io.Writer
	errs io.Writer
	file string
	properties
}

const (
	ConfigVarName = "SHELP_CONFIG"
	RootVarName   = "SHELP_ROOT"
)

func NewConfig(out, err io.Writer) Config {
	return Config{outs: out, errs: err}
}

func (c *Config) LoadConfig(path string) error {
	type pathArg struct {
		path string
		fail bool
	}
	pathArgs := []pathArg{
		{path, true},
		{os.Getenv(ConfigVarName), true},
		{c.defaultConfigPath(), false},
	}

	for _, pa := range pathArgs {
		if pa.path == "" {
			continue
		}
		_, err := os.Stat(pa.path)
		if !os.IsNotExist(err) {
			// Load config
			file, err := os.Open(pa.path)
			if err != nil {
				fmt.Fprintf(c.errs, "Error! Can't open config file: %s. Error = %v\n", pa.path, err)
				return err
			}

			decoder := yaml.NewDecoder(file)
			if err := decoder.Decode(&c.properties); err != nil {
				fmt.Fprintf(c.errs, "Error! Load config failed. path: %s, error: %v\n", pa.path, err)
				return err
			}

			c.file = pa.path
			break
		} else if pa.fail {
			fmt.Fprintf(c.errs, "Error! File not found: %s\n", pa.path)
			return err
		}
	}

	return nil
}

// IsLoaded means whether config file is loaded or not
func (c *Config) IsLoaded() bool {
	return c.file != ""
}

// File returns path of loaded config file
func (c *Config) File() string {
	return c.file
}

// RootPath wraps rootPath, has side effect
func (c *Config) RootPath() string {
	if c.Path.Root == "" {
		c.Path.Root = c.rootPath()
	}
	return c.Path.Root
}

func (c *Config) PackagePath() string {
	return filepath.Join(c.RootPath(), "packages")
}

func (c *Config) BinPath() string {
	return filepath.Join(c.RootPath(), "bin")
}

func (c *Config) TempPath() string {
	return filepath.Join(c.RootPath(), "tmp")
}

// Same as RootPath(), but don't set c.Path.Root
func (c *Config) rootPath() string {
	if c.Path.Root == "" {
		if path := os.Getenv(RootVarName); path != "" {
			return path
		}
		return defaultRootPath()
	}
	return c.Path.Root
}

func (c *Config) defaultConfigPath() string {
	return filepath.Join(c.rootPath(), "config.yml")
}

func defaultRootPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Can't get HomeDir! Error: %v", err))
	}
	return filepath.Join(home, ".shelp")
}
