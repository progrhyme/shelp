[![release](https://badgen.net/github/release/progrhyme/shelp)](https://github.com/progrhyme/shelp/releases)
[![go-test](https://github.com/progrhyme/shelp/workflows/go-test/badge.svg)](https://github.com/progrhyme/shelp/actions?query=workflow%3Ago-test)

# shelp

`shelp` is a Git-based package manager for shell scripts written in Go.

# What is this for?

With `shelp`, you can do the followings:

- Install any git repositories reachable with `git` command and organize them under `$SHELP_ROOT` directory. `shelp` treat them as _packages_
- Add any executable files in a _package_ into `$PATH` when `shelp` installs it
- Load any shell script in a _package_ easily by `include` function bundled in `shelp`
- Manage what _packages_ to be installed and how by the configuration file; and install them at once
- Specify any git branch or tag or commit hash for a _package_ to install

# System Requirements

- OS: Linux or macOS
- `git` command

Supported Shells:

- Bash, Zsh and most POSIX compatible shells
- fish shell

# Documentation

Full documentation is here: https://go-shelp.netlify.app/ .

# Installation

There are several ways to install `shelp` :

- [Homebrew](https://brew.sh/) or [Linuxbrew](https://docs.brew.sh/Homebrew-on-Linux) (using Tap)
- Download from GitHub releases
- go get (go command is required)

Choose one which is suitable for you.

## Homebrew (Linuxbrew)

```sh
brew tap progrhyme/tap
brew install shelp
```

## Download from Releases

Download latest binary from [GitHub Releases](https://github.com/progrhyme/shelp/releases)
and put it under one directory in `$PATH` entries.

Let's see typical commands to achieve this:

```sh
bin=/usr/local/bin  # Change to your favorite path
version=0.6.0       # Make sure this is the latest
os=darwin           # or "linux" is supported
curl -Lo $bin/shelp "https://github.com/progrhyme/shelp/releases/download/v${version}/shelp_${version}_${os}_x86_64"
chmod +x $bin/shelp
```

## go get

Run the following:

```sh
go get github.com/progrhyme/shelp
```

# Usage

Go to [Documentation site](https://go-shelp.netlify.app/).

# Alternatives

There are other tools to manage shell scripts in modular way.  
Pick up some of them here.

 Software | Supported Shells
----------|------------------
 [basherpm/basher](https://github.com/basherpm/basher) | Bash, Zsh, fish shell
 [zplug](https://github.com/zplug/zplug) | Zsh
 [bpkg](https://www.bpkg.sh/) | Bash
 [jorgebucaran/fisher](https://github.com/jorgebucaran/fisher) | fish shell

# Special Thanks

**basher** inspired me to implement some features in this tool.

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
