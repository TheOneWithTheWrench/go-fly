package internal

const zshInitSnippet = `fly() {
  case "$1" in
    -h|--help)
      command fly "$@"
      return $?
      ;;
  esac

  local dest
  dest="$(command fly "$@")" || return 1
  if [[ -n "$dest" && -d "$dest" ]]; then
    cd "$dest"
  fi
}

_fly_track() {
  command fly track >/dev/null 2>&1
}

if [[ -z "$` + ChildEnvVar + `" ]]; then
  autoload -U add-zsh-hook
  if ! add-zsh-hook -L chpwd 2>/dev/null | command grep -q "_fly_track"; then
    add-zsh-hook chpwd _fly_track
  fi

  _fly_track
fi
`

func ZshInitSnippet() string {
	return zshInitSnippet
}
