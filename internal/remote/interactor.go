package remote

import (
	"context"
	"fmt"
	"sgit/github"
	"sgit/internal/cmd"
	"sgit/internal/logging"
	"strings"
	"sync"
)

type (
	Interactor struct {
		logger *logging.Logger
		github *github.Client
	}
)

func NewInteractor(l *logging.Logger, token, username string) *Interactor {
	return &Interactor{
		l,
		github.NewClient(token, username),
	}
}

func (i Interactor) GetRepos(ctx context.Context) ([]cmd.RemoteRepository, error) {
	repos, err := i.github.GetAllRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("github.GetAllRepos failed: %w", err)
	}

	var wg sync.WaitGroup
	results := make(chan cmd.RemoteRepository, len(repos))
	for _, repo := range repos {
		wg.Add(1)
		go func(r github.Repository, results chan<- cmd.RemoteRepository) {
			defer wg.Done()
			results <- i.normalizeAndDetermineLanguage(ctx, r)
		}(repo, results)
	}

	wg.Wait()
	close(results)

	rv := make([]cmd.RemoteRepository, 0)
	for repo := range results {
		rv = append(rv, repo)
	}

	return rv, nil
}

func (i Interactor) normalizeAndDetermineLanguage(ctx context.Context, r github.Repository) cmd.RemoteRepository {
	p := strings.Split(r.FullName, "/")
	name := p[len(p)-1]

	normalized := cmd.RemoteRepository{
		Repository: cmd.Repository{
			Name:   name,
			SshUrl: r.SshUrl,
		},
		Fork: r.Fork,
	}

	if r.Owner != nil {
		normalized.Owner = r.Owner.Login
	}

	lang, err := i.github.GetPrimaryLanguageForRepo(ctx, name)
	if err != nil {
		i.logger.Error(err, "github.GetPrimaryLanguageForRepo failed", "name", name)
	}

	normalized.Language = strings.ToLower(lang)
	return normalized
}
