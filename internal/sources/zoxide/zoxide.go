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
	"time"

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

type SourceOption func(*sourceOptions) error

type sourceOptions struct {
	lister Lister
}

type refreshableLister interface {
	Refresh(context.Context) error
}

func WithLister(lister Lister) SourceOption {
	return func(opts *sourceOptions) error {
		if lister == nil {
			return fmt.Errorf("zoxide lister required")
		}

		opts.lister = lister
		return nil
	}
}

func New(optionFuncs ...SourceOption) (*Source, error) {
	defaultOpts := sourceOptions{}

	for _, option := range optionFuncs {
		if err := option(&defaultOpts); err != nil {
			return nil, err
		}
	}

	if defaultOpts.lister == nil {
		lister, err := NewCommandLister(defaultRunner())
		if err != nil {
			return nil, err
		}

		defaultOpts.lister = lister
	}

	return &Source{lister: defaultOpts.lister}, nil
}

func (s *Source) Load(ctx context.Context, query string) ([]internal.Candidate, error) {
	matches, err := s.lister.List(ctx)
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

func (s *Source) Refresh(ctx context.Context) error {
	if refresher, ok := s.lister.(refreshableLister); ok {
		return refresher.Refresh(ctx)
	}

	_, err := s.lister.List(ctx)
	return err
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
	runner         internal.Runner
	shell          string
	store          *Store
	now            func() time.Time
	backendTimeout time.Duration
}

type CommandListerOption func(*CommandLister)

func WithBackendTimeout(d time.Duration) CommandListerOption {
	return func(l *CommandLister) {
		l.backendTimeout = d
	}
}

func NewCommandLister(runner internal.Runner, opts ...CommandListerOption) (*CommandLister, error) {
	if runner == nil {
		return nil, fmt.Errorf("runner required")
	}

	shell := os.Getenv("SHELL")
	if shell == "" {
		shell = "sh"
	}

	store, err := NewStore()
	if err != nil {
		return nil, err
	}

	l := &CommandLister{runner: runner, shell: shell, store: store, now: time.Now, backendTimeout: defaultBackendTimeout}
	for _, opt := range opts {
		opt(l)
	}

	return l, nil
}

func (l *CommandLister) List(ctx context.Context) ([]Match, error) {
	cache, exists, err := l.store.Load()
	if err != nil {
		cache = Cache{}
		exists = false
	}

	if !ShouldRefresh(cache, exists) {
		return cache.Matches, nil
	}

	result, backend, err := l.listWithFallback(ctx, cache.Backend)
	if err != nil {
		return nil, wrapListError(err)
	}

	filtered := filterGitMatches(result)
	err = l.store.Save(Cache{
		FetchedAt: l.now().UTC(),
		Backend:   backend,
		Matches:   filtered,
	})
	if err != nil {
		return result, nil
	}

	return result, nil
}

func (l *CommandLister) Refresh(ctx context.Context) error {
	cache, _, err := l.store.Load()
	if err != nil {
		cache = Cache{}
	}

	result, backend, err := l.listWithFallback(ctx, cache.Backend)
	if err != nil {
		return wrapListError(err)
	}

	filtered := filterGitMatches(result)
	err = l.store.Save(Cache{
		FetchedAt: l.now().UTC(),
		Backend:   backend,
		Matches:   filtered,
	})
	if err != nil {
		return err
	}

	return nil
}

const (
	backendZoxide         = "zoxide"
	backendZ              = "z"
	backendShell          = "shell"
	defaultBackendTimeout = 5 * time.Second
)

func (l *CommandLister) listWithFallback(ctx context.Context, preferredBackend string) ([]Match, string, error) {
	backends := backendOrder(preferredBackend)
	var lastErr error

	for _, backend := range backends {
		backendCtx, cancel := context.WithTimeout(ctx, l.backendTimeout)
		result, err := l.runBackend(backendCtx, backend)
		cancel()
		if err == nil {
			return result, backend, nil
		}
		if !isCommandNotFound(err) {
			return nil, "", err
		}

		lastErr = err
	}

	if lastErr == nil {
		return nil, "", fmt.Errorf("no zoxide backends configured")
	}

	return nil, "", lastErr
}

func (l *CommandLister) runBackend(ctx context.Context, backend string) ([]Match, error) {
	switch backend {
	case backendZoxide:
		return l.runAndParse(ctx, "zoxide", "query", "--list", "--score")
	case backendZ:
		return l.runAndParse(ctx, "z", "-l")
	case backendShell:
		return l.runAndParse(ctx, l.shell, "-ic", "z -l")
	default:
		return nil, fmt.Errorf("unsupported zoxide backend: %s", backend)
	}
}

func backendOrder(preferred string) []string {
	order := []string{backendZoxide, backendZ, backendShell}
	if !isSupportedBackend(preferred) {
		return order
	}

	prioritized := []string{preferred}
	for _, backend := range order {
		if backend == preferred {
			continue
		}
		prioritized = append(prioritized, backend)
	}

	return prioritized
}

func isSupportedBackend(backend string) bool {
	switch backend {
	case backendZoxide, backendZ, backendShell:
		return true
	default:
		return false
	}
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

func defaultRunner() internal.Runner {
	return internal.RunnerFunc(func(ctx context.Context, name string, args ...string) ([]byte, error) {
		cmd := exec.CommandContext(ctx, name, args...)
		output, err := cmd.CombinedOutput()
		if err != nil {
			return output, fmt.Errorf("run %s %v: %w: %s", name, args, err, output)
		}

		return output, nil
	})
}

func wrapListError(err error) error {
	return fmt.Errorf("list zoxide entries: %w", err)
}

func filterGitMatches(matches []Match) []Match {
	filtered := make([]Match, 0, len(matches))

	for _, match := range matches {
		valid, err := internal.CheckDestination(match.Path)
		if err != nil || !valid {
			continue
		}

		filtered = append(filtered, match)
	}

	return filtered
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
