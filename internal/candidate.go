package internal

type Candidate struct {
	Signals  map[string]float64
	Meta     map[string]string
	resolver Source
}

const (
	CandidateMetaLabel    = "label"
	CandidateMetaKind     = "kind"
	CandidateMetaPath     = "path"
	CandidateMetaName     = "name"
	CandidateMetaFullName = "full_name"
	CandidateMetaSSHURL   = "ssh_url"
	CandidateMetaSource   = "source"

	CandidateSourceLocal  = "local"
	CandidateSourceRemote = "remote"
	CandidateSourceZoxide = "zoxide"

	CandidateSignalZoxideScore = "source.zoxide.score"
)
