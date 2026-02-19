package internal

import (
	"fmt"
	"io"
	"strings"
)

const Usage = "usage: fly [query] | fly init | fly refresh | fly track"

type CliDependencies struct {
	Init    func(io.Writer) error
	Refresh func() error
	Prune   func() error
	Track   func() error
	Query   func(string, io.Writer) error
}

func Run(args []string, stdout io.Writer, stderr io.Writer, deps CliDependencies) error {
	if len(args) < 2 {
		if deps.Query == nil {
			return fmt.Errorf("query handler not configured")
		}
		return deps.Query("", stdout)
	}

	if args[1] == "-h" || args[1] == "--help" {
		_, err := fmt.Fprintln(stdout, Usage)
		return err
	}

	switch args[1] {
	case "init":
		if deps.Init == nil {
			return fmt.Errorf("init handler not configured")
		}
		return deps.Init(stdout)
	case "refresh":
		if deps.Refresh == nil {
			return fmt.Errorf("refresh handler not configured")
		}
		return deps.Refresh()
	case "_prune":
		if deps.Prune == nil {
			return fmt.Errorf("prune handler not configured")
		}
		return deps.Prune()
	case "track":
		if deps.Track == nil {
			return fmt.Errorf("track handler not configured")
		}
		return deps.Track()
	default:
		if deps.Query == nil {
			return fmt.Errorf("query handler not configured")
		}
		query := ""
		if len(args) > 1 {
			query = strings.Join(args[1:], " ")
		}
		return deps.Query(query, stdout)
	}
}
