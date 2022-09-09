goshodo
=======

[![Test Status](https://github.com/Songmu/goshodo/workflows/test/badge.svg?branch=main)][actions]
[![MIT License](https://img.shields.io/github/license/Songmu/goshodo)][license]
[![PkgGoDev](https://pkg.go.dev/badge/github.com/Songmu/goshodo)][PkgGoDev]

[actions]: https://github.com/Songmu/goshodo/actions?workflow=test
[license]: https://github.com/Songmu/goshodo/blob/main/LICENSE
[PkgGoDev]: https://pkg.go.dev/github.com/Songmu/goshodo

goshodo is a CLI tool for shodo (https://shodo.ink) in Go

## Synopsis

```console
% export SHODO_API_TOKEN=...
% export SHODO_API_ROOT=https://...
% goshodo lint testdata/demo.md
Linting...
3:11 もしかしてAI
    飛行機の欠便があり、運行（→ 運航）状況が変わった。 バ
6:5 もしかしてAI
    ません。  これが私で（→ の）自己紹介です。  こ
...
```

## Installation

```console
# Install the latest version. (Install it into ./bin/ by default).
% curl -sfL https://raw.githubusercontent.com/Songmu/goshodo/main/install.sh | sh -s

# Specify installation directory ($(go env GOPATH)/bin/) and version.
% curl -sfL https://raw.githubusercontent.com/Songmu/goshodo/main/install.sh | sh -s -- -b $(go env GOPATH)/bin [vX.Y.Z]

# In alpine linux (as it does not come with curl by default)
% wget -O - -q https://raw.githubusercontent.com/Songmu/goshodo/main/install.sh | sh -s [vX.Y.Z]

# go install
% go install github.com/Songmu/goshodo/cmd/goshodo@latest
```

## See Also
- https://shodo.ink
- https://github.com/zenproducts/shodo-python/

## Author

[Songmu](https://github.com/Songmu)
