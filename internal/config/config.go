package config

import (
	"fmt"
	"os"
	"path/filepath"
)

type Config struct{}

func NewConfig() Config {
	return *new(Config)
}

func (c *Config) RootPath() string {
	path := os.Getenv("CLAFT_ROOT")
	if path == "" {
		path = c.defaultRootPath()
	}
	return path
}

func (c *Config) PackagePath() string {
	return filepath.Join(c.RootPath(), "packages")
}

func (c *Config) BinPath() string {
	return filepath.Join(c.RootPath(), "bin")
}

func (c *Config) defaultRootPath() string {
	home, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("Can't get HomeDir! Error: %v", err))
	}
	return filepath.Join(home, ".claft")
}
