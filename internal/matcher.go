package internal

import (
	"strings"
)

// Filter pre-filters local entries for explicit queries.
// The goal is a higher chance of a direct hit when the user types a query.
// Empty queries skip this and show the full list for fuzzy browsing.
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

// FilterRemote pre-filters remote repos for explicit queries.
// This keeps the candidate set tight so a query is likely to resolve to one repo.
// Empty queries skip this and show the full list for fuzzy browsing.
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
