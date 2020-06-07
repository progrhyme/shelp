# shelp

`shelp` is a Git-based package manager for shell scripts written in Go.

# Prerequisites

`git` command is needed.

# Installation

// TODO: Release binary.

To enable `shelp` automatically in one's shell, append the following to its
profile (such as `~/.bashrc` or `~/.zshrc`):

```sh
eval "$(shelp init -)"
```

# CLI Usage

```sh
shelp COMMAND [arguments...] [options...]
shelp -h|--help          # Show help
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

# License

The MIT License.

Copyright (c) 2020 IKEDA Kiyoshi.
