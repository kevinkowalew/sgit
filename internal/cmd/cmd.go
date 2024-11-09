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
	UncommittedChanges
	NotGitRepo
	NoRemoteRepo
	IncorrectLanguageParentDirectory
	NotCloned
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
		langsSet                       *set.Set[string]
		statesSet                      *set.Set[string]
		forks                          *bool
	}
)

func (rs RepositoryState) String() string {
	switch rs {
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

func New(
	l *logging.Logger,
	ri RemoteInteractor,
	li LocalInteractor,
	cloneMissingRepos, outputRepos bool,
	langsSet, statesSet *set.Set[string],
	forks *bool,
) *Cmd {
	return &Cmd{
		logger:            l,
		remote:            ri,
		local:             li,
		cloneMissingRepos: cloneMissingRepos,
		outputRepos:       outputRepos,
		langsSet:          langsSet,
		statesSet:         statesSet,
		forks:             forks,
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
		if !c.shouldIncludeRemote(remote) {
			continue
		}

		s := state{
			fork:     remote.Fork,
			fullPath: filepath.Join(c.local.BaseDir(), remote.Language, remote.Name),
		}

		locals, ok := localRepoMap[remote.Repository.Name]
		if !ok {
			if !c.cloneMissingRepos {
				s.repoState = NotCloned
			} else if err := c.local.Clone(remote); err != nil {
				c.logger.Error(err, "local.Clone failed", "name", remote.Repository.Name)
				s.repoState = NotCloned
			} else {
				s.repoState = UpToDate
			}

			if c.statesSet.Size() > 0 && !c.statesSet.Contains(s.repoState.String()) {
				continue
			}

			repoStateMap[remote.Repository] = s
		}

		for _, local := range locals {
			if !c.shouldInclude(local.Repository) {
				continue
			}

			s := state{
				fork:     remote.Fork,
				fullPath: filepath.Join(c.local.BaseDir(), local.Language, remote.Name),
			}

			_, ok := repoStateMap[remote.Repository]
			if ok {
				continue
			}

			if local.Repository.Language != remote.Language {
				s.repoState = IncorrectLanguageParentDirectory
			} else {
				s.repoState = UpToDate
			}

			if c.statesSet.Size() > 0 && !c.statesSet.Contains(s.repoState.String()) {
				continue
			}

			repoStateMap[remote.Repository] = s
		}
	}

	if !c.outputRepos {
		return nil
	}

	if c.forks == nil || !*c.forks {
		for _, local := range localRepos {
			if !c.shouldInclude(local.Repository) {
				continue
			}
			s := state{
				fullPath: filepath.Join(c.local.BaseDir(), local.Language, local.Name),
			}
			_, ok := remoteRepoMap[local.Repository.Name]
			switch {
			case !ok && local.GitRepo:
				s.repoState = NotGitRepo
			case !ok:
				s.repoState = NoRemoteRepo
			case local.UncommitedChanges:
				// TODO: update this to handle
				s.repoState = UncommittedChanges
			default:
				s.repoState = UpToDate
			}

			if c.statesSet.Size() > 0 && !c.statesSet.Contains(s.repoState.String()) {
				continue
			}

			repoStateMap[local.Repository] = s
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
			case UncommittedChanges:
				d = color.New(color.FgYellow, color.Bold)
			case NotCloned:
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

func (c Cmd) shouldIncludeRemote(repo RemoteRepository) bool {
	if !c.shouldInclude(repo.Repository) {
		return false
	}

	if c.forks == nil {
		return true
	}

	if *c.forks && !repo.Fork {
		return false
	} else if !*c.forks && repo.Fork {
		return false
	}

	return true
}

func (c Cmd) shouldInclude(repo Repository) bool {
	if c.langsSet.Size() > 0 {
		return c.langsSet.Contains(repo.Language)
	}

	return true
}
