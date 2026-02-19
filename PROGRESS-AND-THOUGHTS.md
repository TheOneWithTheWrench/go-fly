# fly (go-fly) progress and thoughts

## Pending
- Define next milestone (docs, release, polish).

## In Progress
- GitHub integration complete; validating UX.

## Done
- Draft initial project plan.
- Updated plan with zsh-only integration, absolute-path uniqueness, full remote cache fetch, and TDD note.
- Renamed progress doc to PROGRESS-AND-THOUGHTS.md.
- Initialized Go module.
- Added CLI skeleton at cmd/fly/main.go.
- Added XDG helpers with tests.
- Added atomic file writer for safe persistence.
- Implemented local index store, matcher, git root detection, and CLI routing (TDD).
- Implemented zsh init snippet, track hook, and local query picker.
- Added dependencies (go-fuzzyfinder, testify) and goimports run.
- Ran `go test ./... -race`.
- Updated query flow to prefilter and always open fuzzyfinder with query prefilled.
- Adjusted CLI to allow `fly` with no query (opens full index).
- Added `fly --help` output handling.
- Updated zsh init snippet to guard help flags and only cd into valid directories.
- Added remote cache store with fetched_at timestamp and tests.
- Added GitHub client wrapper with Runner abstraction and tests.
- Implemented `fly refresh` to sync remote cache.
- Integrated remote cache into query flow with background refresh and clone on select.
- Added matcher support for remote repos and refresh staleness tests.
- Added clone destination checks to avoid cloning into existing non-repo directories.
- Filtered remote candidates when a local repo with the same name exists.
- Refactored main logic into internal/app, internal/candidate, internal/clone, and internal/picker.
- Flattened internal packages into a single internal package.
- Dropped AppDependencies in favor of explicit NewApp parameters.
- Inlined App dependencies as direct fields on App.
- Moved nil dependency checks into NewApp and removed internal guard checks.
- Switched app and gh tests to moq-generated mocks.
- Replaced ExecRunner with RunnerFunc and added helper constructors in main.
- Added background prune state storage and hidden _prune flow for local index cleanup.
- Removed missing local selections on query and pruned stale entries.

## Thoughts
- Keep dependencies minimal; use stdlib CLI parsing unless it grows.
- Local index keyed by absolute repo path.
- Remote cache stores full repo list for fast local search.
- Defer GitHub integration until local flow is solid (gather/search/jump).
