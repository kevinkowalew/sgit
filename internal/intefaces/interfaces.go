package interfaces

import "sgit/internal/types"

type GithubClient interface {
	GetPrimaryLanguageForRepo(n string) (string, error)
	GetAllRepos() ([]types.GithubRepository, error)
}

type FilesystemClient interface {
	CreateDirectory(path string) error
	Exists(path string) (bool, error)
}

type GitClient interface {
	CloneRepo(r types.GithubRepository, path string) error
	HasLocalChanges(path string) (bool, error)
	PushLocalChanges(path string) error
	StashLocalChanges(path string) error
	ResetLocalChanges(path string) error
	PullLatestChanges(path string) error
}
