# go-fly

[![CI](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/ci.yml/badge.svg)](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/ci.yml)
[![Release](https://img.shields.io/github/actions/workflow/status/TheOneWithTheWrench/go-fly/release.yml?label=release)](https://github.com/TheOneWithTheWrench/go-fly/actions/workflows/release.yml)
[![GitHub release](https://img.shields.io/github/v/release/TheOneWithTheWrench/go-fly)](https://github.com/TheOneWithTheWrench/go-fly/releases)
[![Go version](https://img.shields.io/github/go-mod/go-version/TheOneWithTheWrench/go-fly)](https://go.dev/doc/devel/release)
[![Go Report Card](https://goreportcard.com/badge/github.com/TheOneWithTheWrench/go-fly)](https://goreportcard.com/report/github.com/TheOneWithTheWrench/go-fly)

go-fly is a fast CLI to jump to local git repos by name. It can also discover and clone remote GitHub repos.

**Status:** Under development. Early and still evolving.

## Install & Setup (zsh)
Install with Homebrew:

```bash
brew install TheOneWithTheWrench/tap/go-fly
```

Add this to your `~/.zshrc`:

```zsh
eval "$(fly init)"
```

Optional but recommended on first install:

```bash
fly refresh
```

`fly refresh` fetches your GitHub repos and caches them locally so they show up immediately when you search. Running it once up front avoids a cold start and makes the first `fly <query>` feel instant.

## Usage
- `fly`: open the full list.
- `fly <query>`: narrow the list and jump to a repo.
- `fly -c <query>`: force remote clones into the current working directory for this run.
- `fly refresh`: fetch GitHub repos and update the local cache.

## Config
`fly` reads config from `~/.config/fly/config.toml` (or `$FLY_CONFIG` if set).

To set a default directory for remote clones:

```toml
version = 1

[clone]
default_directory = "/absolute/path/to/repos"
```

When `clone.default_directory` is set, remote repos clone there by default.
Use `fly -c <query>` to override that and clone into the current working directory.
