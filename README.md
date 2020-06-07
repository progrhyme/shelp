# shelp

`shelp` is a Git-based package manager for shell scripts written in Go.

# Prerequisites

`git` command is needed.

# Installation
## Install Binary

There are two ways to install `shelp` binary:

1. Download from GitHub releases
1. go get

Let's see in details.

### Download from releases

Download latest binary from GitHub [releases](https://github.com/progrhyme/shelp/releases)
and put it under one directory in `$PATH` entries.

Let's see typical commands to achieve this:

```sh
bin=/usr/local/bin  # Change to your favorite path
version=0.1.0       # Make sure this is the latest
os=darwin           # or "linux" is supported
curl -Lo $bin/shelp "https://github.com/progrhyme/shelp/releases/download/v${version}/shelp_${version}_${os}_x86_64"
chmod +x $bin/shelp
```

### go get

Run the following:

```sh
go get github.com/progrhyme/shelp
```

## Enable in Shell

To enable `shelp` automatically in your shell, append the following to your
profile script (such as `~/.bashrc` or `~/.zshrc`):

```sh
eval "$(shelp init -)"
```

# Configuration

You can set `SHELP_ROOT` environment variable before using `shelp`.  
The default value is `~/.shelp`.

# Quickstart
## Install Packages

Limitation:

- In shelp v0.1.0, only installation from github.com via https scheme is supported.

Command Syntax:

```sh
shelp install <account>/<repository>
```

For example, the following command installs https://github.com/bats-core/bats-core
into `$SHELP_ROOT/packages/bats-core` directory.

```sh
shelp install bats-core/bats-core
```

Then, you can run `bats` command in the package.

## Shell Function

The command `eval "$(shelp init -)"` loads a shell function `include`.

Function Usage:

```sh
include <package> <script-path>
```

This will load `<script-path>` in `<package>` by `.` shell built-in function.

For example, suppose you have installed https://github.com/ohmyzsh/ohmyzsh by shelp.  
Then, the following command load oh-my-zsh.sh on your current shell:

```sh
include ohmyzsh oh-my-zsh.sh
```

# CLI Usage

```sh
shelp COMMAND [arguments...] [options...]
shelp -h|--help          # Show general help
shelp COMMAND -h|--help  # Show help for COMMAND
shelp -v|--version       # Show CLI version
```

## Available Commands

```sh
shelp init       # Initialize shelp for shell environment
shelp install    # Install a package
shelp add        # Alias of "install"
shelp remove     # Uninstall a package
shelp uninstall  # Alias of "remove"
shelp list       # List installed packages
shelp destroy    # Delete all materials including packages
```

# Alternatives

- [basherpm/basher](https://github.com/basherpm/basher)

# Special Thanks

**basher** inspired me to implement some features in this tool.

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
