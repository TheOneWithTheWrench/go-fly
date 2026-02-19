package internal

import (
	"fmt"
	"path"
	"path/filepath"
	"strings"
)

type Kind int

const (
	KindLocal Kind = iota
	KindRemote
)

type Candidate struct {
	Kind   Kind
	Local  Entry
	Remote Repo
}

func Build(locals []Entry, remotes []Repo) []Candidate {
	localNames := make(map[string]struct{}, len(locals))
	candidates := make([]Candidate, 0, len(locals)+len(remotes))

	for _, entry := range locals {
		name := entry.Name
		if name == "" {
			name = filepath.Base(entry.Path)
		}
		if name != "" {
			localNames[strings.ToLower(name)] = struct{}{}
		}
		candidates = append(candidates, Candidate{Kind: KindLocal, Local: entry})
	}

	for _, repo := range remotes {
		remoteName := repo.Name
		if remoteName == "" {
			remoteName = path.Base(repo.FullName)
		}
		if remoteName != "" {
			if _, exists := localNames[strings.ToLower(remoteName)]; exists {
				continue
			}
		}
		if repo.FullName != "" {
			if _, exists := localNames[strings.ToLower(path.Base(repo.FullName))]; exists {
				continue
			}
		}
		candidates = append(candidates, Candidate{Kind: KindRemote, Remote: repo})
	}

	return candidates
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
