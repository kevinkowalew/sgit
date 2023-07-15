package cmd

import (
	"fmt"
	"os"
	"sgit/internal/filesystem"
	"sgit/internal/git"
	"sgit/internal/github"
	"sgit/internal/intefaces"
	"sgit/internal/types"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
)

type refreshCommand struct {
	github     interfaces.GithubClient
	git        interfaces.GitClient
	filesystem interfaces.FilesystemClient
	targetDir  string
}

func NewRefreshCommand() (*refreshCommand, error) {
	token := lookupEnv("GITHUB_TOKEN")
	org := lookupEnv("GITHUB_ORG")
	targetDir := lookupEnv("CODE_HOME_DIR")

	githubClient := github.NewClient(token, org)
	gitClient := git.NewClient()
	filesystemClient := filesystem.NewClient()
	return newRefreshCommand(githubClient, gitClient, filesystemClient, targetDir), nil
}

func lookupEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic("Unset environment variable: " + key)
	}

	return val
}

func newRefreshCommand(github interfaces.GithubClient, git interfaces.GitClient, filesystem interfaces.FilesystemClient, targetDir string) *refreshCommand {
	return &refreshCommand{github, git, filesystem, targetDir}
}

func (c refreshCommand) refresh(repo types.GithubRepository) (*repositoryState, error) {
	lang, err := c.github.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s/%s/%s", c.targetDir, strings.ToLower(lang), repo.Name())
	exists, err := c.filesystem.Exists(path)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = c.filesystem.CreateDirectory(path)
		if err != nil {
			return nil, err
		}

		err = c.git.CloneRepo(repo, path)
		if err != nil {
			return nil, err
		}
		return &repositoryState{repo, false, path}, nil
	}

	hasLocalChanges, err := c.git.HasLocalChanges(path)
	if err != nil {
		return nil, err
	}

	if !hasLocalChanges {
		err = c.git.PullLatestChanges(path)
		if err != nil {
			return nil, err
		}
		return &repositoryState{repo, false, path}, nil
	}

	return &repositoryState{repo, true, path}, nil
}

type repositoryState struct {
	r               types.GithubRepository
	hasLocalChanges bool
	path            string
}

func (c refreshCommand) Run() error {
	// TODO: implement reported failures here
	var wg sync.WaitGroup
	ac := make(chan struct {
		*repositoryState
		error
	})

	allRepos, err := c.github.GetAllRepos()
	if err != nil {
		return err
	}
	for _, r := range allRepos {
		wg.Add(1)
		go func(r types.GithubRepository, ac chan struct {
			*repositoryState
			error
		}) {
			defer wg.Done()
			state, err := c.refresh(r)
			ac <- struct {
				*repositoryState
				error
			}{state, err}
		}(r, ac)
	}

	go func() {
		for rs := range ac {
			if !rs.hasLocalChanges {
				continue
			}

			p := promptui.Select{
				Label: rs.r.Name() + " has local changes",
				Items: []string{"Skip", "Push", "Stash", "Reset"},
			}

			_, a, err := p.Run()
			if err != nil {
				panic(err)
			}

			if a == "Skip" {
				continue
			}

			// TODO: figure this out
			if a == "Push" {
				err = c.git.PushLocalChanges(rs.path)
			} else if a == "Stash" {
				err = c.git.StashLocalChanges(rs.path)
			} else if a == "Reset" {
				err = c.git.ResetLocalChanges(rs.path)
			}
			err = c.git.PullLatestChanges(rs.path)
		}
	}()

	wg.Wait()
	close(ac)
	return nil
}
