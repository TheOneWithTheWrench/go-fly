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
`fly` is configurable.

- Default config path: `~/.config/fly/config.toml`
- Override path with: `FLY_CONFIG=/path/to/config.toml`
- If no config file exists, `fly` uses built-in defaults.

Full default config (equivalent to running with no config file):

```toml
version = 1

[clone]
default_directory = ""
group_by_owner = false

[sources]
enabled = ["zoxide", "remote"]

[picker]
title = ""
prompt_marker = "> "
window_position = "bottom"
matcher = "orderedchars"
sorter = "signal_minipick"
```

Available options:

- `version`: config schema version. Currently `1`.
- `clone.default_directory`: default destination for remote clones. Empty means current working directory. `fly -c <query>` always clones into current working directory. Use an absolute path or `~/...`.
- `clone.group_by_owner`: when `true`, remote clones use `<base>/<owner>/<repo>` instead of `<base>/<repo>`.

  Example:

  ```toml
  [clone]
  default_directory = "~/projects"
  # or: default_directory = "/home/your-user/projects"
  ```
- `sources.enabled`:
  - `zoxide`: include repos discovered from your zoxide database (with zoxide score signal)
  - `remote`: include cached GitHub repos (cloned on selection)
  - `local`: include repos tracked directly by `fly track`
- `picker.title`: optional picker title (empty to hide)
- `picker.prompt_marker`: prompt prefix for the query input
- `picker.window_position`:
  - `top`: prompt at top, list shown in natural order
  - `bottom`: prompt at bottom, list shown in reverse order
- `picker.matcher`:
  - `orderedchars`: query chars must appear in order (not necessarily contiguous)
  - `substring`: query must appear as a contiguous, case-insensitive substring
- `picker.sorter`:
  - `minipick`: mini.pick-inspired sorting by tighter match width first, then earlier match start
  - `signal_minipick`: prioritize items with zoxide score signal, then fall back to `minipick`
