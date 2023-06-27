package cmd

import (
	"fmt"
	"os"
	"sgit/internal/github_client"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
)

type RefreshCommand struct {
	githubClient *github_client.GithubClient
	targetDir    string
}

func NewRefreshCommand() (*RefreshCommand, error) {
	token := lookupEnv("GITHUB_TOKEN")
	org := lookupEnv("GITHUB_ORG")
	targetDir := lookupEnv("CODE_HOME_DIR")
	githubClient, err := github_client.NewGithubClient(token, org, targetDir)
	if err != nil {
		return nil, err
	}
	return &RefreshCommand{githubClient, targetDir}, nil
}

func lookupEnv(key string) string {
	val, ok := os.LookupEnv(key)
	if !ok {
		panic("Unset environment variable: " + key)
	}

	return val
}

func (c RefreshCommand) refresh(repo github_client.Repository) repositoryState {
	lang, err := c.githubClient.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("%s/%s/%s", c.targetDir, strings.ToLower(lang), repo.Name())
	exists, err := exists(path)
	if err != nil {
		panic(err)
	}

	if !exists {
		c.githubClient.CloneRepo(repo)
		return repositoryState{repo, false, path}
	}

	if !c.githubClient.HasLocalChanges(path) {
		c.githubClient.PullLatestChanges(path)
		return repositoryState{repo, false, path}
	}

	return repositoryState{repo, true, path}
}

type repositoryState struct {
	r     github_client.Repository
	stale bool
	path  string
}

func (c RefreshCommand) Run() {
	var wg sync.WaitGroup
	ac := make(chan repositoryState)
	for _, r := range c.githubClient.GetAllRepos() {
		wg.Add(1)
		go func(r github_client.Repository, ac chan repositoryState) {
			defer wg.Done()
			ac <- c.refresh(r)
		}(r, ac)
	}

	go func() {
		for rs := range ac {
			if !rs.stale {
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
				c.githubClient.PushLocalChanges(rs.path)
			} else if a == "Stash" {
				c.githubClient.StashLocalChanges(rs.path)
			} else if a == "Reset" {
				c.githubClient.ResetLocalChanges(rs.path)
			}
			c.githubClient.PullLatestChanges(rs.path)
		}
	}()

	wg.Wait()
	close(ac)
}

func exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
