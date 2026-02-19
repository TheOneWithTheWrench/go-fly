# go-fly

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

## To be developed
- Homebrew install support.
