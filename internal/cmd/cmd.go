package cmd

import (
	"errors"
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
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return nil, errors.New("Unset environment variable: GITHUB_TOKEN")
	}

	org, ok := os.LookupEnv("GITHUB_ORG")
	if !ok {
		return nil, errors.New("Unset environment variable: GITHUB_ORG")
	}

	targetDir, ok := os.LookupEnv("CODE_HOME_DIR")
	if !ok {
		return nil, errors.New("Unset environment variable: CODE_HOME_DIR")
	}

	githubClient := github.NewClient(token, org)
	gitClient := git.NewClient()
	filesystemClient := filesystem.NewClient()
	return newRefreshCommand(githubClient, gitClient, filesystemClient, targetDir), nil
}

func newRefreshCommand(github interfaces.GithubClient, git interfaces.GitClient, filesystem interfaces.FilesystemClient, targetDir string) *refreshCommand {
	return &refreshCommand{github, git, filesystem, targetDir}
}

func (c *refreshCommand) refresh(repo types.GithubRepository) (*repositoryState, error) {
	lang, err := c.github.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		return nil, err
	}

	baseDir := fmt.Sprintf("%s/%s", c.targetDir, strings.ToLower(lang))
	path := fmt.Sprintf("%s/%s", baseDir, repo.Name())
	exists, err := c.filesystem.Exists(path)
	if err != nil {
		return nil, err
	}

	if !exists {
		err = c.filesystem.CreateDirectory(path)
		if err != nil {
			return nil, err
		}
		err = c.git.CloneRepo(repo, baseDir)
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

func (c *refreshCommand) Run() error {
	var wg sync.WaitGroup
	allRepos, err := c.github.GetAllRepos()
	if err != nil {
		return err
	}

	ac := make(chan struct {
		*repositoryState
		error
	})

	for _, r := range allRepos {
		_, err := c.refresh(r)
		if err != nil {
			fmt.Println(err.Error())
		}
		wg.Add(1)
		go func(r types.GithubRepository, ac chan struct {
			*repositoryState
			error
		}) {
			defer wg.Done()
			repoState, err := c.refresh(r)
			ac <- struct {
				*repositoryState
				error
			}{repoState, err}
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

			if a == "Push" {
				err = c.git.PushLocalChanges(rs.path)
			} else if a == "Stash" {
				err = c.git.StashLocalChanges(rs.path)
			} else if a == "Reset" {
				err = c.git.ResetLocalChanges(rs.path)
			}

			if err != nil {
				fmt.Println("Error: " + err.Error())
				continue
			}

			err = c.git.PullLatestChanges(rs.path)
			if err != nil {
				fmt.Println("Error: " + err.Error())
			}
		}
	}()

	wg.Wait()
	close(ac)
	return nil
}
