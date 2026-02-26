package internal

type Repo struct {
	Name     string `json:"name"`
	FullName string `json:"full_name"`
	SSHURL   string `json:"ssh_url"`
}
