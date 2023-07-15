package interfaces

import "internal/types"

type GithubClient interface {
	CloneRepo(r types.Repository)
	GetPrimaryLanguageForRepo(n string) (string, error)
	GetAllRepos() []types.Repository
}

type FilesystemClient interface {
	CreateDirectory(path string) error
}

type GitClient interface {
	HasLocalChanges(path string) (bool, error)
	PushLocalChanges(path string) error
	StashLocalChanges(path string) error
	ResetLocalChanges(path string) error
	PullLatestChanges(path string) error
}
