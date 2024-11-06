package main

import (
	"context"
	"os"
	"sgit/internal/cmd"
	"sgit/internal/local"
	"sgit/internal/logging"
	"sgit/internal/remote"
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

	ctx := context.Background()
	l := logging.New()
	li := local.NewInterfactor(l, targetDir)
	ri := remote.NewInteractor(l, token, username)
	cmd := cmd.New(l, ri, li)

	if err := cmd.Run(ctx); err != nil {
		panic(err)
	}
}
