package zoxide

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/TheOneWithTheWrench/go-fly/internal"
)

type Match struct {
	Path  string
	Score float64
}

type Lister interface {
	List(context.Context) ([]Match, error)
}

type Source struct {
	lister Lister
}

func New(lister Lister) (*Source, error) {
	if lister == nil {
		return nil, fmt.Errorf("zoxide lister required")
	}

	return &Source{lister: lister}, nil
}

func (s *Source) Load(query string) ([]internal.Candidate, error) {
	matches, err := s.lister.List(context.Background())
	if err != nil {
		return nil, err
	}
	if len(matches) == 0 {
		return nil, internal.ErrNoReposTracked
	}

	filtered := filterMatches(query, matches)
	candidates := make([]internal.Candidate, 0, len(filtered))

	for _, match := range filtered {
		candidate, ok := localCandidate(match)
		if !ok {
			continue
		}

		candidates = append(candidates, candidate)
	}

	return candidates, nil
}

func localCandidate(match Match) (internal.Candidate, bool) {
	valid, err := internal.CheckDestination(match.Path)
	if err != nil || !valid {
		return internal.Candidate{}, false
	}

	entry := internal.Entry{
		Name: filepath.Base(match.Path),
		Path: match.Path,
	}

	return internal.Candidate{
		Signals: map[string]float64{internal.CandidateSignalZoxideScore: match.Score},
		Meta: map[string]string{
			internal.CandidateMetaLabel:  fmt.Sprintf("%s (%s) [zoxide]", entry.Name, entry.Path),
			internal.CandidateMetaSource: internal.CandidateSourceZoxide,
			internal.CandidateMetaName:   entry.Name,
			internal.CandidateMetaPath:   entry.Path,
		},
	}, true
}

func (s *Source) Resolve(candidate internal.Candidate) (string, error) {
	if candidate.Meta[internal.CandidateMetaSource] != internal.CandidateSourceZoxide {
		return "", internal.ErrUnsupportedCandidate
	}

	path := candidate.Meta[internal.CandidateMetaPath]
	if path == "" {
		return "", internal.ErrUnsupportedCandidate
	}

	valid, err := internal.CheckDestination(path)
	if err != nil || !valid {
		return "", fmt.Errorf("repo no longer exists: %s", path)
	}

	return path, nil
}

func filterMatches(query string, matches []Match) []Match {
	if strings.TrimSpace(query) == "" {
		return matches
	}

	needle := strings.ToLower(query)
	filtered := make([]Match, 0, len(matches))

	for _, match := range matches {
		path := strings.ToLower(match.Path)
		name := strings.ToLower(filepath.Base(match.Path))
		if strings.Contains(path, needle) || strings.Contains(name, needle) {
			filtered = append(filtered, match)
		}
	}

	return filtered
}

type CommandLister struct {
	runner internal.Runner
	shell  string
}

func NewCommandLister(runner internal.Runner) (*CommandLister, error) {
	if runner == nil {
		return nil, fmt.Errorf("runner required")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	return &CommandLister{runner: runner, shell: shell}, nil
}

func (l *CommandLister) List(ctx context.Context) ([]Match, error) {
	result, err := l.runAndParse(ctx, "zoxide", "query", "--list", "--score")
	if err == nil {
		return result, nil
	}
	if !isCommandNotFound(err) {
		return nil, wrapListError(err)
	}

	result, err = l.runAndParse(ctx, "z", "-l")
	if err == nil {
		return result, nil
	}
	if !isCommandNotFound(err) {
		return nil, wrapListError(err)
	}

	result, err = l.runAndParse(ctx, l.shell, "-ic", "z -l")
	if err == nil {
		return result, nil
	}

	return nil, wrapListError(err)
}

func (l *CommandLister) runAndParse(ctx context.Context, name string, args ...string) ([]Match, error) {
	data, err := l.runner.Run(ctx, name, args...)
	matches := parseOutput(string(data))
	if len(matches) > 0 {
		return matches, nil
	}
	if err != nil {
		return nil, err
	}

	return nil, nil
}

func wrapListError(err error) error {
	return fmt.Errorf("list zoxide entries: %w", err)
}

func parseOutput(output string) []Match {
	lines := strings.Split(output, "\n")
	result := make([]Match, 0, len(lines))
	pathOnly := make([]string, 0, len(lines))

	for _, line := range lines {
		score, path, ok := parseScoredLine(line)
		if ok {
			result = append(result, Match{Path: path, Score: score})
			continue
		}

		path, ok = parsePathLine(line)
		if ok {
			pathOnly = append(pathOnly, path)
		}
	}

	if len(result) > 0 {
		return result
	}

	if len(pathOnly) == 0 {
		return nil
	}

	result = make([]Match, len(pathOnly))
	for i, path := range pathOnly {
		score := float64(len(pathOnly) - i)
		result[i] = Match{Path: path, Score: score}
	}

	return result
}

func parseScoredLine(line string) (float64, string, bool) {
	line = strings.TrimSpace(line)
	if line == "" {
		return 0, "", false
	}

	sep := strings.IndexAny(line, " \t")
	if sep < 0 {
		return 0, "", false
	}

	scorePart := strings.TrimSpace(line[:sep])
	path := strings.TrimSpace(line[sep+1:])
	if path == "" {
		return 0, "", false
	}

	score, err := strconv.ParseFloat(scorePart, 64)
	if err != nil {
		return 0, "", false
	}

	return score, path, true
}

func parsePathLine(line string) (string, bool) {
	path := strings.TrimSpace(line)
	if path == "" {
		return "", false
	}

	if strings.HasPrefix(path, "/") || strings.HasPrefix(path, "./") || strings.HasPrefix(path, "../") || strings.HasPrefix(path, "~") {
		return path, true
	}

	return "", false
}

func isCommandNotFound(err error) bool {
	if errors.Is(err, exec.ErrNotFound) {
		return true
	}

	return strings.Contains(err.Error(), "executable file not found")
}
