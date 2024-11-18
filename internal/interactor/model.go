package interactor

const (
	UpToDate State = iota + 1
	UncommittedChanges
	NotGitRepo
	NoRemoteRepo
	IncorrectLanguageParentDirectory
	NotCloned
)

type Repo struct {
	Name, Language, SshUrl, Path     string
	Fork, GitRepo, UncommitedChanges bool
	Owner                            string
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
