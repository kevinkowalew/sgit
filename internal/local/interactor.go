package local

import (
	"fmt"
	"path/filepath"
	"sgit/filesystem"
	"sgit/git"
	"sgit/internal/cmd"
	"sgit/internal/logging"
	"strings"
	"sync"
)

type (
	Filesystem interface {
		CreateDirectory(relativePath string) error
		Exists(relativePath string) (bool, error)
		ListDirectories() ([]string, error)
	}

	Git interface {
		GetSshUrl(fullPath string) (string, error)
		Clone(repo cmd.RemoteRepository, path string) error
		HasUncommittedChanges(path string) (bool, error)
	}

	Interactor struct {
		baseDir string
		logger  *logging.Logger
		fs      Filesystem
		git     Git
	}
)

func NewInterfactor(l *logging.Logger, baseDir string) *Interactor {
	fs := filesystem.NewClient(baseDir)
	git := git.New()
	return &Interactor{baseDir, l, fs, git}
}

func (i Interactor) GetRepos() ([]cmd.LocalRepository, error) {
	dirs, err := i.fs.ListDirectories()
	if err != nil {
		return nil, fmt.Errorf("filesystem.ListDirectories failed: %w", err)
	}

	var wg sync.WaitGroup
	results := make(chan cmd.LocalRepository, len(dirs))
	for _, dir := range dirs {
		wg.Add(1)
		go func(dir string, results chan<- cmd.LocalRepository) {
			defer wg.Done()
			results <- i.normalize(dir)
		}(dir, results)
	}

	wg.Wait()
	close(results)

	rv := make([]cmd.LocalRepository, 0)
	for repo := range results {
		rv = append(rv, repo)
	}

	return rv, nil
}

func (i Interactor) normalize(dir string) cmd.LocalRepository {
	if strings.HasSuffix(dir, "/") {
		dir = dir[0 : len(dir)-1]
	}

	p := strings.Split(dir, "/")
	name := p[len(p)-1]
	lang := ""
	if len(p) > 1 {
		lang = p[len(p)-2]
	}

	repo := cmd.LocalRepository{
		Repository: cmd.Repository{
			Name:     name,
			Language: lang,
		},
	}

	fullPath := filepath.Join(dir, ".git/")
	isGitRepo, err := i.fs.Exists(fullPath)
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
	repo.Repository.SshUrl = sshUrl

	uncommittedChanges, err := i.git.HasUncommittedChanges(dir)
	if err != nil {
		i.logger.Error(err, "git.HasUncommittedChanges failed", "name", name, "lang", lang)
	}
	repo.UncommitedChanges = uncommittedChanges

	return repo
}

func (li Interactor) Clone(repo cmd.RemoteRepository) error {
	parent := filepath.Join(li.baseDir, repo.Language)
	if err := li.fs.CreateDirectory(parent); err != nil {
		return fmt.Errorf("fs.CreateDirectory failed: %w", err)
	}

	return li.git.Clone(repo, parent)
}

func (li Interactor) BaseDir() string {
	return li.baseDir
}
