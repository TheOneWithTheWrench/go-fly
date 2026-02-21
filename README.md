# go-fly

[![CI](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/ci.yml/badge.svg)](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/TheOneWithTheWrench/go-fly/release.yml?label=release)](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/TheOneWithTheWrench/go-fly)](https://github.com/TheOneWithTheWrench/go-fly/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/TheOneWithTheWrench/go-fly)](https://go.dev/doc/devel/release)
[![Go Report Card](https://goreportcard.com/badge/github.com/TheOneWithTheWrench/go-fly)](https://goreportcard.com/report/github.com/TheOneWithTheWrench/go-fly)

go-fly is a fast CLI to jump to local git repos by name. It can also find and clone remote repos via GitHub.

**Status:** Under development. Early and still evolving.

## Setup (zsh)
Add this to your `~/.zshrc`:

```zsh
eval "$(fly init)"
```

## Usage
- `fly <query>`: open a fuzzy picker and jump to a repo.
- `fly refresh`: refresh the remote repo cache.
- `fly`: open the full list.

## Install (Homebrew)
```bash
brew install TheOneWithTheWrench/tap/go-fly
```

## To be developed
- Homebrew install support.
