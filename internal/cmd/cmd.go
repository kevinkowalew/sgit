package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sgit/internal/logging"
	"strings"
)

const (
	UpToDate RepositoryState = iota + 1
	UpToDateWithLocalChanges
	HasMergeConflicts
	HasUncommittedChanges
	HasUncommittedChangesAndBehindUpstream
	NoUpstreamRepo
	FailedToClone
	IncorrectLanguageParentDirectory
)

type (
	RepositoryState int

	Repository struct {
		Name, Language, SshUrl string
	}

	RemoteRepository struct {
		Repository
		Fork  bool
		Owner string
	}

	LocalRepository struct {
		Repository
		GitRepo, UncommitedChanges bool
	}

	LocalInteractor interface {
		GetRepos() ([]LocalRepository, error)
		Clone(repo RemoteRepository) error
		BaseDir() string
	}

	RemoteInteractor interface {
		GetRepos(ctx context.Context) ([]RemoteRepository, error)
	}

	cmd struct {
		logger *logging.Logger
		local  LocalInteractor
		remote RemoteInteractor
	}
)

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
	case FailedToClone:
		return "FailedToClone"
	case IncorrectLanguageParentDirectory:
		return "IncorrectLanguageParentDirectory"
	}

	return ""
}

func New(l *logging.Logger, ri RemoteInteractor, li LocalInteractor) *cmd {
	return &cmd{
		logger: l,
		remote: ri,
		local:  li,
	}
}

func (c *cmd) Run(ctx context.Context) error {
	remoteRepos, err := c.remote.GetRepos(ctx)
	if err != nil {
		return fmt.Errorf("remote.GetRepos failed: %w", err)
	}

	remoteRepoMap := make(map[string]RemoteRepository, 0)
	for _, repo := range remoteRepos {
		remoteRepoMap[repo.Name] = repo
	}

	localRepos, err := c.local.GetRepos()
	if err != nil {
		return fmt.Errorf("local.GetRepos failed: %w", err)
	}

	localRepoMap := make(map[string][]LocalRepository, 0)
	for _, repo := range localRepos {
		e, _ := localRepoMap[repo.Name]
		localRepoMap[repo.Name] = append(e, repo)
	}

	type state struct {
		fork      bool
		fullPath  string
		repoState RepositoryState
	}

	repoStateMap := make(map[Repository]state, 0)
	for _, remote := range remoteRepos {
		locals, ok := localRepoMap[remote.Repository.Name]
		if !ok {
			if err := c.local.Clone(remote); err != nil {
				c.logger.Error(err, "local.Clone failed", "name", remote.Repository.Name)
				repoStateMap[remote.Repository] = state{
					fork:      remote.Fork,
					fullPath:  filepath.Join(c.local.BaseDir(), remote.Language, remote.Name),
					repoState: FailedToClone,
				}
			} else {
				repoStateMap[remote.Repository] = state{
					fork:      remote.Fork,
					fullPath:  filepath.Join(c.local.BaseDir(), remote.Language, remote.Name),
					repoState: UpToDate,
				}
			}
		}

		for _, local := range locals {
			_, ok := repoStateMap[remote.Repository]
			if ok {
				continue
			}

			if local.Repository.Language != remote.Language {
				repoStateMap[remote.Repository] = state{
					fork:      remote.Fork,
					fullPath:  filepath.Join(c.local.BaseDir(), local.Language, remote.Name),
					repoState: IncorrectLanguageParentDirectory,
				}
			} else {
				repoStateMap[remote.Repository] = state{
					fork:      remote.Fork,
					fullPath:  filepath.Join(c.local.BaseDir(), local.Language, remote.Name),
					repoState: UpToDate,
				}
			}
		}
	}

	for _, local := range localRepos {
		_, ok := remoteRepoMap[local.Repository.Name]
		if !ok {
			repoStateMap[local.Repository] = state{
				fullPath:  filepath.Join(c.local.BaseDir(), local.Language, local.Name),
				repoState: NoUpstreamRepo,
			}
		}

		if local.UncommitedChanges {
			repoStateMap[local.Repository] = state{
				fullPath:  filepath.Join(c.local.BaseDir(), local.Language, local.Name),
				repoState: HasUncommittedChanges,
			}
		}
	}

	langToRepo := make(map[string][]state, 0)

	for repo, state := range repoStateMap {
		e, _ := langToRepo[repo.Language]
		langToRepo[repo.Language] = append(e, state)
	}

	for lang, states := range langToRepo {
		for _, state := range states {
			fork := "Fork"
			if !state.fork {
				fork = "Not" + fork
			}
			parts := []string{lang, state.fullPath, state.repoState.String(), fork}
			fmt.Println(strings.Join(parts, " "))
		}
	}

	return nil
}
