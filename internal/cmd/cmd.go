package cmd

import (
	"context"
	"fmt"
	"sgit/internal/logging"

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
		Clone(ctx context.Context, repo Repository) error
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

	localRepoMap := make(map[string]LocalRepository, 0)
	for _, repo := range localRepos {
		localRepoMap[repo.Name] = repo
	}

	repoStateMap := make(map[Repository]RepositoryState, 0)

	for _, repo := range remoteRepos {
		_, ok := localRepoMap[repo.Name]
		if !ok {
			if err := c.local.Clone(ctx, repo.Repository); err != nil {
				c.logger.Error(err, "local.Clone failed", "name", repo.Name)
				repoStateMap[repo.Repository] = FailedToClone
				continue
			}
		}

		repoStateMap[repo.Repository] = UpToDate
	}

	for _, repo := range localRepos {
		_, ok := remoteRepoMap[repo.Name]
		if !ok {
			repoStateMap[repo.Repository] = NoUpstreamRepo
		}

		if repo.UncommitedChanges {
			repoStateMap[repo.Repository] = HasUncommittedChanges
		}
	}

	type repoStatePair struct {
		repo  Repository
		state RepositoryState
	}
	langToRepo := make(map[string][]repoStatePair, 0)

	for repo, state := range repoStateMap {
		e, _ := langToRepo[repo.Language]
		langToRepo[repo.Language] = append(e, repoStatePair{repo, state})
	}

	rainbow := []color.Attribute{
		color.FgBlue, color.FgMagenta, color.FgCyan,
	}

	i := 0
	for lang, pairs := range langToRepo {
		for _, pair := range pairs {
			d := color.New(
				rainbow[i%(len(rainbow)-1)],
				color.Bold,
			)
			d.Print(lang + " ")

			d = color.New(color.FgWhite)
			d.Print(pair.repo.Name + " ")

			switch pair.state {
			case UpToDate:
				d = color.New(color.FgGreen, color.Bold)
				d.Println(pair.state.String())
			case HasUncommittedChanges:
				d = color.New(color.FgYellow, color.Bold)
				d.Println(pair.state.String())
			default:
				d = color.New(color.FgRed, color.Bold)
				d.Println(pair.state.String())
			}

		}
		i += 1
	}

	return nil
}
