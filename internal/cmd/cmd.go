package cmd

import (
	"errors"
	"fmt"
	"sgit/internal/github_client"
	"os"
	"strings"
	"sync"

	"github.com/manifoldco/promptui"
)

type RefreshCommand struct {
	Gc *github_client.GithubClient
}

func NewRefreshCommand() (*RefreshCommand, error) {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		return nil, errors.New("Unset environment variable: GITHUB_TOKEN")
	}

	org, ok := os.LookupEnv("GITHUB_ORG")
	if !ok {
		return nil, errors.New("Unset environment variable: GITHUB_ORG")
	}

	gc, err := github_client.NewGithubClient(token, org)
	if err != nil {
		return nil , err
	}
	return &RefreshCommand{gc}, nil
}

func (c RefreshCommand) refresh(repo github_client.Repository) repositoryState {
	lang, err := c.Gc.GetPrimaryLanguageForRepo(repo.Name())
	if err != nil {
		panic(err)
	}

	h, err := os.UserHomeDir()
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("%s/code/%s/%s/", h, strings.ToLower(lang), repo.Name())
	exists, err := exists(path)
	if err != nil {
		panic(err)
	}

	if !exists {
		c.Gc.CloneRepo(repo)
		return repositoryState{repo, false, path}
	}

	if !c.Gc.HasLocalChanges(path) {
		c.Gc.PullLatestChanges(path)
		return repositoryState{repo, false, path}
	}

	return repositoryState{repo, true, path}
}

type repositoryState struct {
	r github_client.Repository
	stale bool
	path string
}

func (c RefreshCommand) Run() {
	var wg sync.WaitGroup 
	ac := make(chan repositoryState)
	for _, r := range c.Gc.GetAllRepos() {
		wg.Add(1)
		go func(r github_client.Repository, ac chan repositoryState) {
			defer wg.Done()
			ac <- c.refresh(r)
		}(r, ac)
	}

	go func() {
	   for rs := range ac {
		   if !rs.stale  {
			   continue
		   }

		   p := promptui.Select{
			   Label: rs.r.Name() + " has local changes",
			   Items: []string{"Skip", "Push", "Stash", "Reset"},
		   }

		   _, a, err :=  p.Run()
		   if err != nil {
			   panic(err)
		   }

		   if a == "Skip" {
			   continue
		   } 

		   if a == "Push" {
			   c.Gc.PushLocalChanges(rs.path)
		   } else if a == "Stash" {
			   c.Gc.StashLocalChanges(rs.path)
		   } else  if a == "Reset" {
			   c.Gc.ResetLocalChanges(rs.path)
		   }
		   c.Gc.PullLatestChanges(rs.path)
	   }
	}()

	wg.Wait()
	close(ac)
}

func exists(path string) (bool, error) {
    _, err := os.Stat(path)
    if err == nil { return true, nil }
    if os.IsNotExist(err) { return false, nil }
    return false, err
}
