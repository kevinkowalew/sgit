package main

import (
	"os"
	"sgit/filesystem"
	"sgit/git"
	"sgit/github"
	"sgit/internal/cmd"
	"sgit/internal/logging"
	"sgit/tui"
)

func main() {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		panic("Unset environment variable: GITHUB_TOKEN")
	}

	username, ok := os.LookupEnv("GITHUB_USERNAME")
	if !ok {
		panic("Unset environment variable: GITHUB_ORG")
	}

	targetDir, ok := os.LookupEnv("CODE_HOME_DIR")
	if !ok {
		panic("Unset environment variable: CODE_HOME_DIR")
	}

	logger := logging.New()
	github := github.NewClient(token, username)
	git := git.NewClient()
	filesystem := filesystem.NewClient()
	tui := tui.New()

	rc := cmd.NewRefreshCommand(logger, github, git, filesystem, tui, targetDir)
	if err := rc.Run(); err != nil {
		panic(err)
	}
}
