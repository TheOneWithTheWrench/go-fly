package internal

import (
	"fmt"
	"io"
	"strings"
)

type CliDependencies struct {
	Init    func(io.Writer) error
	Refresh func() error
	Prune   func() error
	Track   func() error
	Query   func(string, io.Writer, ...QueryOption) error
}

const (
	CloneToCWDFlag     = "-c"
	CloneToCWDLongFlag = "--cwd"
	Usage              = "usage: fly [query] | fly -c [query] | fly init | fly refresh | fly track"
)

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
		return deps.Init(stdout)
	case "refresh":
		return deps.Refresh()
	case "_prune":
		return deps.Prune()
	case "track":
		return deps.Track()
	default:
		var (
			queryParts   = args[1:]
			queryOptions = make([]QueryOption, 0, 1)
		)

		if isCloneToCWDFlag(args[1]) {
			queryOptions = append(queryOptions, WithForceCloneToCWD())
			queryParts = args[2:]
		}

		query := strings.Join(queryParts, " ")
		return deps.Query(query, stdout, queryOptions...)
	}
}

func isCloneToCWDFlag(value string) bool {
	return value == CloneToCWDFlag || value == CloneToCWDLongFlag
}
