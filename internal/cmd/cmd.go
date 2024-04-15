package cmd

import (
	"fmt"
	"sgit/internal/logging"
	"strings"
	"sync"
)

const (
	UpToDate RepositoryState = iota + 1
	UpToDateWithLocalChanges
	HasMergeConflicts
	HasUncommittedChanges
	HasUncommittedChangesAndBehindUpstream
	NoUpstreamRepo
)

type (
	RepositoryState int

	GithubRepository struct {
		FullName string `json:"full_name"`
		SshUrl   string `json:"ssh_url"`
	}

	Github interface {
		GetPrimaryLanguageForRepo(name string) (string, error)
		GetAllRepos() ([]GithubRepository, error)
		GetCommitHash(name, branch string) (string, error)
	}

	Filesystem interface {
		CreateDirectory(path string) error
		Exists(path string) (bool, error)
		ListDirectories(path string, depth int) ([]string, error)
	}

	TUI interface {
		Handle(repository GithubRepository, state RepositoryState) error
	}

	Git interface {
		GetBranchName(path string) (string, error)
		CloneRepo(r GithubRepository, path string) error
		HasUncommittedChanges(path string) (bool, error)
		HasMergeConflicts(path string) (bool, error)
		PullLatest(path string) error
		GetCommitHashes(path string) ([]string, error)
	}

	refreshCommand struct {
		logger     *logging.Logger
		github     Github
		git        Git
		filesystem Filesystem
		tui        TUI
		targetDir  string
		repoStates map[string]RepositoryState
		mutex      sync.RWMutex
	}
)

func (r GithubRepository) Name() string {
	p := strings.Split(r.FullName, "/")
	return p[len(p)-1]
}

func NewRefreshCommand(logger *logging.Logger, github Github, git Git, filesystem Filesystem, tui TUI, targetDir string) *refreshCommand {
	return &refreshCommand{
		logger:     logger,
		github:     github,
		git:        git,
		filesystem: filesystem,
		tui:        tui,
		targetDir:  targetDir,
		repoStates: make(map[string]RepositoryState),
	}
}

func (c *refreshCommand) remoteToLocalRepoRefresh(repo GithubRepository) error {
	lang, err := c.github.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		return c.wrapError(err, "github.GetPrimaryLanguageForRepo failed for "+repo.Name())
	}

	baseDir := fmt.Sprintf("%s/%s", c.targetDir, strings.ToLower(lang))
	path := fmt.Sprintf("%s/%s", baseDir, repo.Name())

	exists, err := c.filesystem.Exists(path)
	if err != nil {
		return c.wrapError(err, "filesystem.Exists failed for "+repo.Name())
	}

	if !exists {
		err = c.filesystem.CreateDirectory(path)
		if err != nil {
			return c.wrapError(err, "filesystem.CreateDirectory failed for "+path)
		}

		err = c.git.CloneRepo(repo, baseDir)
		if err != nil {
			return c.wrapError(err, "git.CloneRepo failed for "+repo.Name())
		}

		return c.tui.Handle(repo, UpToDate)
	}

	branch, err := c.git.GetBranchName(path)
	if err != nil {
		return c.wrapError(err, "git.GetBranchName failed for "+repo.Name())
	}

	//TODO: update these to be branch aware
	hasLocalChanges, err := c.git.HasUncommittedChanges(path)
	if err != nil {
		return c.wrapError(err, "git.HasUncommittedChanges failed for "+repo.Name())
	}

	if !hasLocalChanges {
		err = c.git.PullLatest(path)
		if err != nil {
			return c.wrapError(err, "git.PullLatestChanges failed for "+repo.Name())
		}

		hasMergeConflicts, err := c.git.HasMergeConflicts(path)
		if err != nil {
			return c.wrapError(err, "git.HasMergeConflicts failed for "+repo.Name())
		}

		if hasMergeConflicts {
			return c.tui.Handle(repo, HasMergeConflicts)
		} else {
			return c.tui.Handle(repo, UpToDate)
		}
	}

	upstreamCommitHash, err := c.github.GetCommitHash(repo.Name(), branch)
	if err != nil {
		return c.wrapError(err, "github.GetLatestCommitHash failed for repo.Name()")
	}

	localCommitHashes, err := c.git.GetCommitHashes(path)
	if err != nil {
		return c.wrapError(err, "git.GetCurrentCommitHash failed for repo.Name()")
	}

	for _, localHashCommitHash := range localCommitHashes {
		if upstreamCommitHash == localHashCommitHash {
			return c.tui.Handle(repo, HasUncommittedChanges)
		}
	}

	return c.tui.Handle(repo, HasUncommittedChangesAndBehindUpstream)
}

func (c *refreshCommand) localToRemoteRepoRefresh(path string) {
	p := strings.Split(path, "/")
	name := p[len(p)-1]

	_, ok := c.getRepoState(name)
	if !ok {
		c.setRepoState(name, NoUpstreamRepo)
	}
}

func (c *refreshCommand) setRepoState(repoName string, state RepositoryState) {
	c.mutex.Lock()
	c.repoStates[repoName] = state
	c.mutex.Unlock()
}

func (c *refreshCommand) getRepoState(repoName string) (RepositoryState, bool) {
	c.mutex.RLock()
	defer c.mutex.Unlock()
	value, ok := c.repoStates[repoName]
	return value, ok
}

func (c *refreshCommand) wrapError(err error, operation string) error {
	return fmt.Errorf("%s: %s", operation, err.Error())
}

func (c *refreshCommand) Run() []error {
	errs := []error{}
	if e := c.remoteToLocalRefresh(); e != nil {
		errs = append(errs, e...)
	}

	return errs
}

func (c *refreshCommand) remoteToLocalRefresh() []error {
	allRepos, err := c.github.GetAllRepos()
	if err != nil {
		return []error{err}
	}

	errs := []error{}
	for _, r := range allRepos {
		if err := c.remoteToLocalRepoRefresh(r); err != nil {
			wErr := c.wrapError(err, "cmd.remoteToLocalRepoRefresh failed for "+r.Name())
			errs = append(errs, wErr)
		}
	}
	return errs
}

func (c *refreshCommand) localToRemoteRefresh() error {
	localRepos, err := c.filesystem.ListDirectories(c.targetDir, 2)
	if err != nil {
		return err
	}

	for _, r := range localRepos {
		c.localToRemoteRepoRefresh(r)
	}
	return nil
}
