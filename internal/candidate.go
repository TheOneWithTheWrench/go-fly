package internal

import "fmt"

type Kind string

const (
	KindLocal  Kind = "local"
	KindRemote Kind = "remote"
)

type Candidate struct {
	Kind   Kind
	Local  Entry
	Remote Repo
}

func CandidateLabel(selected Candidate) string {
	if selected.Kind == KindRemote {
		return remoteLabel(selected.Remote)
	}

	return entryLabel(selected.Local)
}

func entryLabel(entry Entry) string {
	if entry.Name == "" {
		return entry.Path
	}

	return fmt.Sprintf("%s (%s)", entry.Name, entry.Path)
}

func remoteLabel(repo Repo) string {
	label := repo.FullName
	if label == "" {
		label = repo.Name
	}
	return fmt.Sprintf("%s (remote)", label)
}
