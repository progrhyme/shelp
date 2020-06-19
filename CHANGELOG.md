## 0.5.3 (2020-06-19)

Enhance: ([#7](https://github.com/progrhyme/shelp/pull/7))

- (install, bundle) Support to specify commit hash of repository on installation

## 0.5.2 (2020-06-18)

Enhance: ([#6](https://github.com/progrhyme/shelp/pull/6))

- (bundle) Re-install packages when their specs change; i.e. `from` or `at` values in config file

## 0.5.1 (2020-06-16)

Feature: ([#5](https://github.com/progrhyme/shelp/pull/5))

- Add `prune` subcommand to remove packages not configured in config file

Other:

- Refactor a lot

## 0.5.0 (2020-06-14)

Features: ([#4](https://github.com/progrhyme/shelp/pull/4))

- Support YAML config file specified by `-c|--config` option or `SHELP_CONFIG` environment variable
- Add `bundle` subcommand to install packages configured in the YAML

## 0.4.0 (2020-06-13)

Features: ([#3](https://github.com/progrhyme/shelp/pull/3))

- Add `outdated` subcommand to show packages which have updates
- (upgrade) Upgrade all packages with no argument
- (install) Support all protocols available on git-clone command

Other: ([#3](https://github.com/progrhyme/shelp/pull/3))

- (Testing/go) Add tests for some of typical sequential CLI tasks

## 0.3.0 (2020-06-13)

Features: ([#2](https://github.com/progrhyme/shelp/pull/2))

- Add `link` subcommand: pseudo installation from local filesystems; creating symbolic link to
original resource
- (install) Enable to install from any git hosting sites which provides HTTPS protocol: e.g.
bitbucket.org, gitlab.com
- (install) Enable to specify git branch or tag to clone by adding `@<branch-or-tag>` suffix to the
argument

## 0.2.0 (2020-06-11)

Features: ([#1](https://github.com/progrhyme/shelp/pull/1))

- Add `upgrade` subcommand to upgrade installed package
- (install) Enable to set alias of package

Improve: ([#1](https://github.com/progrhyme/shelp/pull/1))

- (install) Continue creating symlinks with errors in the middle

Other: ([#1](https://github.com/progrhyme/shelp/pull/1))

- (Testing) Add some go test codes
- (CI) Add CI of testing using GitHub Actions

## 0.1.0 (2020-06-07)

Initial release.
