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
