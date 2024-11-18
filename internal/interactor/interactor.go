package interactor

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"sgit/filesystem"
	"sgit/git"
	"sgit/github"
	"sgit/internal/logging"
	"strings"
	"sync"
)

type Interactor struct {
	logger            *logging.Logger
	github            *github.Github
	filesystem        *filesystem.Filesystem
	git               *git.Git
	username, baseDir string
}

func New() *Interactor {
	baseDir := os.Getenv("CODE_HOME_DIR")
	username := os.Getenv("GITHUB_USERNAME")
	return &Interactor{
		logging.New(),
		github.New(os.Getenv("GITHUB_TOKEN"), username),
		filesystem.New(baseDir),
		git.New(),
		username,
		baseDir,
	}
}

// GetRepos returns a map of programing langauges to a list of RepositoryState Pair
func (i Interactor) GetRepoStates(ctx context.Context, filter Filter) (map[string][]RepoStatePair, error) {
	remoteRepos, err := i.getRemoteRepos(ctx, filter)
	if err != nil {
		return nil, fmt.Errorf("i.getRemoteRepos failed: %w", err)
	}

	localRepoMap, err := i.getLocalRepoMap(filter)
	if err != nil {
		return nil, fmt.Errorf("local.GetRepos failed: %w", err)
	}

	repoStateMap := make(map[string]RepoStatePair, 0)
	for _, remote := range remoteRepos {
		rsp := RepoStatePair{
			Repo: remote,
		}
		local, ok := localRepoMap[remote.Path]
		if !ok {
			rsp.State = NotCloned

			if filter.ShouldIncludeRepoStatePair(rsp) {
				repoStateMap[remote.Path] = rsp
			}
			continue
		}

		rsp = RepoStatePair{
			Repo: local,
		}

		if local.Language != remote.Language {
			rsp.State = IncorrectLanguageParentDirectory
		} else if local.UncommitedChanges {
			// TODO: update this to handle the UncommitedChanges & UpToDate
			rsp.State = UncommittedChanges
		} else {
			rsp.State = UpToDate
		}

		if filter.ShouldIncludeRepoStatePair(rsp) {
			repoStateMap[remote.Path] = rsp
		}

	}

	for _, local := range localRepoMap {
		if _, ok := remoteRepos[local.Path]; ok {
			continue
		}

		rsp := RepoStatePair{
			Repo: local,
		}

		switch {
		case !local.GitRepo:
			rsp.State = NotGitRepo
		case local.UncommitedChanges:
			// TODO: update this to handle the UncommitedChanges & UpToDate
			rsp.State = UncommittedChanges
		default:
			rsp.State = NoRemoteRepo
		}

		if filter.ShouldIncludeRepoStatePair(rsp) {
			repoStateMap[local.Path] = rsp
		}

	}

	rv := make(map[string][]RepoStatePair, 0)
	for _, rsp := range repoStateMap {
		e, _ := rv[rsp.Language]
		rv[rsp.Language] = append(e, rsp)
	}

	return rv, nil
}

func (i Interactor) getLocalRepoMap(fiter Filter) (map[string]Repo, error) {
	dirs, err := i.filesystem.ListDirectories()
	if err != nil {
		return nil, fmt.Errorf("filesystem.ListDirectories failed: %w", err)
	}

	var wg sync.WaitGroup
	results := make(chan Repo, len(dirs))
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string, results chan<- Repo) {
			defer wg.Done()
			results <- i.normalize(dir)
		}(dir, results)
	}

	wg.Wait()
	close(results)

	rv := make(map[string]Repo, 0)
	for repo := range results {
		rv[repo.Path] = repo
	}

	return rv, nil
}

func (i Interactor) normalize(dir string) Repo {
	if strings.HasSuffix(dir, "/") {
		dir = dir[0 : len(dir)-1]
	}

	p := strings.Split(dir, "/")
	name := p[len(p)-1]
	lang := ""
	if len(p) > 1 {
		lang = p[len(p)-2]
	}

	repo := Repo{
		Name:     name,
		Language: lang,
		Path:     dir,
	}

	fullPath := filepath.Join(dir, ".git/")
	isGitRepo, err := i.filesystem.Exists(fullPath)
	if err != nil {
		i.logger.Error(err, "fs.Exists failed", "name", name, "lang", lang, "path", fullPath)
	}
	repo.GitRepo = isGitRepo

	if !repo.GitRepo {
		return repo
	}

	sshUrl, err := i.git.GetSshUrl(dir)
	if err != nil {
		i.logger.Error(err, "git.GetSshUrl failed", "name", name, "lang", lang)
	}
	repo.SshUrl = sshUrl

	uncommittedChanges, err := i.git.HasUncommittedChanges(dir)
	if err != nil {
		i.logger.Error(err, "git.HasUncommittedChanges failed", "name", name, "lang", lang)
	}
	repo.UncommitedChanges = uncommittedChanges

	return repo
}

func (i Interactor) Clone(lang, url string) error {
	parent := filepath.Join(i.baseDir, lang)
	if err := i.filesystem.CreateDirectory(parent); err != nil {
		return fmt.Errorf("fs.CreateDirectory failed: %w", err)
	}

	return i.git.Clone(url, parent)
}

func (li Interactor) BaseDir() string {
	return li.baseDir
}

func (i Interactor) getRemoteRepos(ctx context.Context, filter Filter) (map[string]Repo, error) {
	repos, err := i.github.GetAllRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("github.GetAllRepos failed: %w", err)
	}

	var wg sync.WaitGroup
	results := make(chan Repo, len(repos))
	for _, repo := range repos {
		wg.Add(1)
		go func(r github.Repository, results chan<- Repo) {
			defer wg.Done()
			results <- i.normalizeAndFetchLanguage(ctx, r)
		}(repo, results)
	}

	wg.Wait()
	close(results)

	rv := make(map[string]Repo, 0)
	for repo := range results {
		if filter.ShouldIncludeRepo(repo) {
			rv[repo.Path] = repo
		}
	}

	return rv, nil
}

func (i Interactor) normalizeAndFetchLanguage(ctx context.Context, r github.Repository) Repo {
	p := strings.Split(r.FullName, "/")
	name := p[len(p)-1]

	normalized := Repo{
		Name:    name,
		SshUrl:  r.SshUrl,
		Fork:    r.Fork,
		GitRepo: true,
	}

	if r.Owner != nil {
		normalized.Owner = r.Owner.Login
	}

	lang, err := i.github.GetPrimaryLanguageForRepo(ctx, i.username, name)
	if err != nil {
		i.logger.Error(err, "github.GetPrimaryLanguageForRepo failed", "name", name)
	} else {
		normalized.Language = strings.ToLower(lang)
		normalized.Path = filepath.Join(i.baseDir, normalized.Language, name)
	}

	return normalized
}

func (i Interactor) GetPrimaryLanguageForRepo(ctx context.Context, owner, name string) (string, error) {
	return i.github.GetPrimaryLanguageForRepo(ctx, owner, name)
}

func (i Interactor) CreateDir(lang string) error {
	path := filepath.Join(i.baseDir, lang)
	return i.filesystem.CreateDirectory(path)
}
