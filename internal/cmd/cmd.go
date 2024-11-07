package cmd

import (
	"context"
	"fmt"
	"path/filepath"
	"sgit/internal/logging"
	"sgit/internal/set"

	"github.com/fatih/color"
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
	NotClonedLocally
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

	Cmd struct {
		logger                         *logging.Logger
		local                          LocalInteractor
		remote                         RemoteInteractor
		cloneMissingRepos, outputRepos bool
		langsFilter                    *set.Set[string]
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
	case NotClonedLocally:
		return "NotClonedLocally"
	}

	return ""
}

func New(l *logging.Logger, ri RemoteInteractor, li LocalInteractor,
	cloneMissingRepos, outputRepos bool, langsFilter *set.Set[string]) *Cmd {
	return &Cmd{
		logger:            l,
		remote:            ri,
		local:             li,
		cloneMissingRepos: cloneMissingRepos,
		outputRepos:       outputRepos,
		langsFilter:       langsFilter,
	}
}

func (c *Cmd) Run(ctx context.Context) error {
	remoteRepos, err := c.remote.GetRepos(ctx)
	if err != nil {
		return fmt.Errorf("remote.GetRepos failed: %w", err)
	}

	remoteRepoMap := make(map[string]RemoteRepository, 0)
	for _, repo := range remoteRepos {
		if c.shouldInclude(repo.Repository) {
			remoteRepoMap[repo.Name] = repo
		}
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
		if !c.shouldInclude(remote.Repository) {
			continue
		}

		locals, ok := localRepoMap[remote.Repository.Name]
		if !ok {
			if !c.cloneMissingRepos {
				repoStateMap[remote.Repository] = state{
					fork:      remote.Fork,
					fullPath:  filepath.Join(c.local.BaseDir(), remote.Language, remote.Name),
					repoState: NotClonedLocally,
				}
				continue
			}

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
			if !c.shouldInclude(local.Repository) {
				continue
			}

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

	if !c.outputRepos {
		return nil
	}

	for _, local := range localRepos {
		if !c.shouldInclude(local.Repository) {
			continue
		}
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

	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}
	i := 0
	for lang, states := range langToRepo {
		for _, state := range states {
			d := color.New(
				rainbow[i%(len(rainbow)-1)],
				color.Bold,
			)
			d.Print(lang + " ")

			d = color.New(color.FgWhite)
			d.Print(state.fullPath + " ")

			switch state.repoState {
			case UpToDate:
				d = color.New(color.FgGreen, color.Bold)
			case HasUncommittedChanges:
				d = color.New(color.FgYellow, color.Bold)
			case NotClonedLocally:
				d = color.New(color.FgHiRed, color.Bold)
			default:
				d = color.New(color.FgRed, color.Bold)
			}
			d.Print(state.repoState.String())

			if state.fork {
				d = color.New(color.FgHiCyan)
				d.Println(" Fork")
			} else {
				d.Println()
			}

		}
		i += 1
	}

	return nil
}

func (c Cmd) shouldInclude(repo Repository) bool {
	if c.langsFilter.Size() > 0 {
		return c.langsFilter.Contains(repo.Language)
	}

	return true
}
