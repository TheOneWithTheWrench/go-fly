package internal

import (
	"strings"
)

func Filter(query string, entries []Entry) []Entry {
	if strings.TrimSpace(query) == "" {
		return entries
	}

	query = strings.ToLower(query)
	filtered := make([]Entry, 0, len(entries))

	for _, entry := range entries {
		name := strings.ToLower(entry.Name)
		path := strings.ToLower(entry.Path)
		if strings.Contains(name, query) || strings.Contains(path, query) {
			filtered = append(filtered, entry)
		}
	}

	return filtered
}

func FilterRemote(query string, repos []Repo) []Repo {
	if strings.TrimSpace(query) == "" {
		return repos
	}

	query = strings.ToLower(query)
	filtered := make([]Repo, 0, len(repos))

	for _, repo := range repos {
		name := strings.ToLower(repo.Name)
		fullName := strings.ToLower(repo.FullName)
		if strings.Contains(name, query) || strings.Contains(fullName, query) {
			filtered = append(filtered, repo)
		}
	}

	return filtered
}
