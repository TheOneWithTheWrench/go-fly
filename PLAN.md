# fly (go-fly) plan

## Goals
- Fast, simple CLI to jump to local git repos by fuzzy name.
- Keep a local index of repos seen over time (from local usage + optional clone).
- Use GitHub CLI (`gh`) to query remote repos for cloning when not local.
- Minimal dependencies; single binary; works well with shells via stdout piping.

## Non-goals (initially)
- No background daemon.
- No automatic filesystem scanning of the whole disk.
- No custom GitHub OAuth; rely on `gh` auth.
- No interactive TUI beyond a simple selector prompt.

## UX / CLI
- `fly <query>`
  - Use a shell function wrapper that calls the Go binary to resolve a path,
    then runs `cd` in the current shell.
  - If local match exists, navigate there.
  - If no local match and a remote match exists, prompt to select + confirm clone, then navigate.
- `fly init`
  - Print zsh-only snippet for `.zshrc` that enables `fly <query>` navigation and automatic tracking.
  - Guard against duplicate hook registration when multiple shells start or config is re-sourced.
- `fly refresh`
  - Refresh cached remote repo list from `gh` on demand.

## Data Sources
- Local index: tracked repos with name + path (keep it minimal); absolute path is the unique key.
- Remote index: repos pulled via `gh` from user + orgs; fetch all accessible repos and cache locally.

### Remote query strategy
- Use `gh repo list` for user and orgs with pagination to fetch everything.
- Cache results locally for fast search and to avoid repeated calls.
- Update cadence: manual `fly refresh`.

## Matching Strategy
- Feed all local candidates into `go-fuzzyfinder` and let it filter.
- If multiple local matches, prompt to select.

## Picker
- Use `go-fuzzyfinder` for interactive selection (same style as fuzzy-clone).
- Inspiration: https://github.com/kjuulh/fuzzy-clone (especially its `go-fuzzyfinder` usage).

## Clone Flow
- If no local match and remote matches exist:
  - Show selector list using `go-fuzzyfinder`.
  - Confirm clone to the current working directory.
  - After clone, add to local index and print path.

## Config
- Config file: `~/.config/fly/config.json` (or XDG default).
- Example fields:
  - `autoTrackOnInit`: enable/disable shell hook suggestion.
  - `maxResults`.

## Storage
- Local index file: `~/.local/share/fly/index.json` (XDG default).
- Remote cache file: `~/.cache/fly/remote.json` (XDG default).
- Keep writes atomic (write tmp then rename).

## MVP Milestones
1. CLI skeleton + config + local index read/write.
2. `fly <query>` for local matching + navigation via shell integration.
3. `gh` remote cache + `fly refresh`.
4. Remote match + clone flow.
5. Shell integration snippet + docs.

## Open Questions / Risks
- Behavior when `gh` is missing or not authenticated.
- Best selector UX (simple numbered prompt vs fzf integration).
- Handling monorepos or nested git repos.

## Testing
- Use TDD for core pieces (matcher, index read/write, cache, CLI argument parsing).
- Integration test for `gh` interaction with test doubles (no network).
