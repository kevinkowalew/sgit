package main

import (
	"os"
	"sgit/internal/cmd"
	"sgit/internal/filesystem"
	"sgit/internal/git"
	"sgit/internal/github"
	"sgit/internal/logging"
	"sgit/internal/tui"
)

func main() {
	token, ok := os.LookupEnv("GITHUB_TOKEN")
	if !ok {
		panic("Unset environment variable: GITHUB_TOKEN")
	}

	org, ok := os.LookupEnv("GITHUB_ORG")
	if !ok {
		panic("Unset environment variable: GITHUB_ORG")
	}

	targetDir, ok := os.LookupEnv("CODE_HOME_DIR")
	if !ok {
		panic("Unset environment variable: CODE_HOME_DIR")
	}

	logger := logging.New()
	github := github.NewClient(token, org)
	git := git.NewClient()
	filesystem := filesystem.NewClient()
	tui := tui.New()

	rc := cmd.NewRefreshCommand(logger, github, git, filesystem, tui, targetDir)
	if err := rc.Run(); err != nil {
		panic(err)
	}
}
