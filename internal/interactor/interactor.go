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
	remoteRepos, err := i.getRemoteRepos(ctx)
	if err != nil {
		return nil, fmt.Errorf("i.getRemoteRepos failed: %w", err)
	}

	localRepoMap, err := i.getLocalRepoMap()
	if err != nil {
		return nil, fmt.Errorf("local.GetRepos failed: %w", err)
	}

	repoStateMap := make(map[string]RepoStatePair, 0)
	for _, remote := range remoteRepos {
		rsp := RepoStatePair{
			Repo: remote,
		}
		local, ok := localRepoMap[remote.Path()]
		if !ok {
			rsp.State = NotCloned
			repoStateMap[remote.Path()] = rsp
			continue
		}

		local.Fork = remote.Fork
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

		repoStateMap[remote.Path()] = rsp
	}

	for _, local := range localRepoMap {
		if _, ok := remoteRepos[local.Path()]; ok {
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

		repoStateMap[local.Path()] = rsp
	}

	var wg sync.WaitGroup
	results := make(chan *RepoStatePair, len(repoStateMap))

	for _, rsp := range repoStateMap {
		wg.Add(1)
		go func(rsp RepoStatePair) {
			defer wg.Done()
			if filter.Include(rsp) {
				results <- &rsp
			} else {
				results <- nil
			}
		}(rsp)

	}

	go func() {
		wg.Wait()
		close(results)
	}()

	rv := make(map[string][]RepoStatePair, 0)
	for rsp := range results {
		if rsp != nil {
			e := rv[rsp.Language]
			rv[rsp.Language] = append(e, *rsp)
		}
	}

	return rv, nil
}

func (i Interactor) getLocalRepoMap() (map[string]Repo, error) {
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

	go func() {
		wg.Wait()
		close(results)
	}()

	rv := make(map[string]Repo, 0)
	for repo := range results {
		rv[repo.Path()] = repo
	}

	return rv, nil
}

func (i Interactor) normalize(dir string) Repo {
	dir = strings.TrimSuffix(dir, "/")

	p := strings.Split(dir, "/")
	name := p[len(p)-1]
	lang := ""
	owner := ""
	if len(p) > 1 {
		lang = p[len(p)-2]
	}

	if len(p) > 2 {
		owner = p[len(p)-3]
	}

	repo := Repo{
		Name:     name,
		Language: lang,
		Owner:    owner,
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
	repo.URL = sshUrl

	uncommittedChanges, err := i.git.HasUncommittedChanges(dir)
	if err != nil {
		i.logger.Error(err, "git.HasUncommittedChanges failed", "name", name, "lang", lang)
	}
	repo.UncommitedChanges = uncommittedChanges

	return repo
}

func (i Interactor) Clone(r Repo) error {
	parent := filepath.Join(i.baseDir, r.Owner, r.Language)
	if err := i.filesystem.CreateDirectory(parent); err != nil {
		return fmt.Errorf("fs.CreateDirectory failed: %w", err)
	}

	return i.git.Clone(r.URL, parent)
}

func (i Interactor) CreateRepo(ctx context.Context, name string, private bool) (*Repo, error) {
	if err := i.github.CreateRepo(ctx, name, private); err != nil {
		return nil, err
	}

	user := os.Getenv("GITHUB_USERNAME")
	return &Repo{
		Name: name,
		// TODO: dynamically detect target language when running create in a repo
		Language: "unknown",
		Owner:    user,
		URL:      fmt.Sprintf("https://github.com/%s/%s", user, name),
	}, nil
}

func (i Interactor) DeleteRemote(ctx context.Context, r Repo) error {
	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid repo: %w", err)
	}
	return i.github.DeleteRepo(ctx, r.Owner, r.Name)
}

func (i Interactor) DeleteLocal(r Repo) error {
	if err := r.Validate(); err != nil {
		return fmt.Errorf("invalid repo: %w", err)
	}
	return i.filesystem.DeleteDir(r.Path())
}

func (i Interactor) Exists(r Repo) (bool, error) {
	return i.filesystem.Exists(r.Path())
}

func (i Interactor) getRemoteRepos(ctx context.Context) (map[string]Repo, error) {
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
		rv[repo.Path()] = repo
	}

	return rv, nil
}

func (i Interactor) normalizeAndFetchLanguage(ctx context.Context, r github.Repository) Repo {
	p := strings.Split(r.FullName, "/")
	name := p[len(p)-1]

	normalized := Repo{
		Name:    name,
		URL:     r.SshUrl,
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
	}

	return normalized
}

func (i Interactor) GetPrimaryLanguageForRepo(ctx context.Context, owner, name string) (string, error) {
	return i.github.GetPrimaryLanguageForRepo(ctx, owner, name)
}
