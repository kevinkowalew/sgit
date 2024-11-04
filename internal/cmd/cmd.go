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
		Fork     bool   `json:"fork"`
		Owner    Owner  `json:"owner"`
		Language string
		State    RepositoryState
	}

	Owner struct {
		Login string `json:"login"`
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
		Handle(repos []GithubRepository) error
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

func (rs RepositoryState) String() string {
	switch rs {
	case UpToDate:
		return "UpToDate"
	case UpToDateWithLocalChanges:
		return "UpToDateWithLocalChanges"
	case HasMergeConflicts:
		return "HasMergeConflicts"
	case HasUncommittedChanges:
		return "HasUncommittedChanges"
	case HasUncommittedChangesAndBehindUpstream:
		return "HasUncommittedChangesAndBehindUpstream"
	case NoUpstreamRepo:
		return "NoUpstreamRepo"
	}

	return ""
}

func NewRefreshCommand(l *logging.Logger, gh Github, g Git, fs Filesystem, tui TUI, targetDir string) *refreshCommand {
	return &refreshCommand{
		logger:     l,
		github:     gh,
		git:        g,
		filesystem: fs,
		tui:        tui,
		targetDir:  targetDir,
		repoStates: make(map[string]RepositoryState),
	}
}

func (c *refreshCommand) remoteToLocalRepoRefresh(repo GithubRepository) (RepositoryState, error) {
	lang, err := c.github.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		return 0, fmt.Errorf("github.GetPrimaryLanguageForRepo failed: %w", err)
	}

	baseDir := fmt.Sprintf("%s/%s", c.targetDir, strings.ToLower(lang))
	path := fmt.Sprintf("%s/%s", baseDir, repo.Name())

	exists, err := c.filesystem.Exists(path)
	if err != nil {
		return 0, fmt.Errorf("filesystem.Exists: %w", err)
	}

	if !exists {
		// create & clone repo if it exist
		err = c.filesystem.CreateDirectory(path)
		if err != nil {
			return 0, fmt.Errorf("filesystem.CreateDirectory failed for "+path+": %w", err)
		}

		err = c.git.CloneRepo(repo, baseDir)
		if err != nil {
			return 0, fmt.Errorf("git.CloneRepo failed: %w")
		}

		return UpToDate, nil
	}

	branch, err := c.git.GetBranchName(path)
	if err != nil {
		return 0, fmt.Errorf("git.GetBranchName failed: %w", err)
	}

	//TODO: update these to be branch aware
	hasLocalChanges, err := c.git.HasUncommittedChanges(path)
	if err != nil {
		return 0, fmt.Errorf("git.HasUncommittedChanges failed: %w", err)
	}

	if !hasLocalChanges {
		err = c.git.PullLatest(path)
		if err != nil {
			return 0, fmt.Errorf("git.PullLatestChanges failed: %w", err)
		}

		hasMergeConflicts, err := c.git.HasMergeConflicts(path)
		if err != nil {
			return 0, fmt.Errorf("git.HasMergeConflicts failed: %w", err)
		}

		if hasMergeConflicts {
			return HasMergeConflicts, nil
		} else {
			return UpToDate, nil
		}
	}

	upstreamCommitHash, err := c.github.GetCommitHash(repo.Name(), branch)
	if err != nil {
		return 0, fmt.Errorf("github.GetLatestCommitHash failed: %w", err)
	}

	localCommitHashes, err := c.git.GetCommitHashes(path)
	if err != nil {
		return 0, fmt.Errorf("git.GetCurrentCommitHash failed: %w", err)
	}

	for _, localHashCommitHash := range localCommitHashes {
		if upstreamCommitHash == localHashCommitHash {
			return HasUncommittedChanges, nil
		}
	}

	return HasUncommittedChangesAndBehindUpstream, nil
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

func (c *refreshCommand) Run() error {
	if err := c.remoteToLocalRefresh(); err != nil {
		return err
	}

	return nil
}

func (c *refreshCommand) remoteToLocalRefresh() error {
	repos, err := c.github.GetAllRepos()
	if err != nil {
		return fmt.Errorf("github.GetAllRepos failed: %w", err)
	}

	results := make(chan GithubRepository, len(repos))

	var wg sync.WaitGroup
	for _, repo := range repos {
		wg.Add(1)
		go func(r GithubRepository, results chan<- GithubRepository) {
			defer wg.Done()
			state, err := c.remoteToLocalRepoRefresh(r)
			if err != nil {
				c.logger.Error(err, "remoteToLocalRefresh failed", "repo_name", r.FullName, "repo_ssh_url", r.SshUrl)
			}
			r.State = state
			results <- r
		}(repo, results)

	}

	wg.Wait()
	close(results)

	toHandle := make([]GithubRepository, 0)
	for repo := range results {
		if repo.Language != "" {
			toHandle = append(toHandle, repo)
		}
	}

	if err := c.tui.Handle(toHandle); err != nil {
		c.logger.Error(err, "tui.Handle failed")
	}

	return nil
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
