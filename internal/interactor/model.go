package interactor

import (
	"os"
	"path/filepath"
)

const (
	UpToDate State = iota + 1
	UncommittedChanges
	NotGitRepo
	NoRemoteRepo
	IncorrectLanguageParentDirectory
	NotCloned
)

type Repo struct {
	Name, Language, Owner, URL       string
	Fork, GitRepo, UncommitedChanges bool
}

func (r Repo) FullName() string {
	return filepath.Join(r.Owner, r.Name)
}

func (r Repo) Path() string {
	baseDir := os.Getenv("CODE_HOME_DIR")
	return filepath.Join(baseDir, r.Owner, r.Language, r.Name)
}

type RepoStatePair struct {
	Repo
	State
}

type State int

func (s State) String() string {
	switch s {
	case UpToDate:
		return "UpToDate"
	case UncommittedChanges:
		return "UncommittedChanges"
	case NotGitRepo:
		return "NotGitRepo"
	case NoRemoteRepo:
		return "NoRemoteRepo"
	case IncorrectLanguageParentDirectory:
		return "IncorrectLanguageParentDirectory"
	case NotCloned:
		return "NotCloned"
	}

	return ""
}
