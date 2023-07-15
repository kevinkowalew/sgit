package types

import "strings"

type Repository struct {
	FullName string `json:"full_name"`
	SshUrl   string `json:"ssh_url"`
}

func (r Repository) Name() string {
	p := strings.Split(r.FullName, "/")
	return p[len(p)-1]
}
