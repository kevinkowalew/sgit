package main

import (
	"context"
	"flag"
	"os"
	"sgit/internal/cmd"
	"sgit/internal/local"
	"sgit/internal/logging"
	"sgit/internal/remote"
	"sgit/internal/set"
	"strings"
)

const (
	ls    = "ls"
	sync  = "sync"
	purge = "purge"
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

	lsCmd := flag.NewFlagSet("ls", flag.ExitOnError)
	lsLangs := lsCmd.String("langs", "", "comma-separated list of languages to target")

	syncCmd := flag.NewFlagSet("sync", flag.ExitOnError)
	syncLangs := syncCmd.String("langs", "", "comma-separated list of languages to target")

	if len(os.Args) < 2 {
		panic("You must specify a subcommand: ls, sync")
	}

	cloneMissingRepos, outputRepos := false, false
	var langMap *set.Set[string]
	switch os.Args[1] {
	case ls:
		lsCmd.Parse(os.Args[2:])
		langMap = createLangsMap(lsLangs)
		outputRepos = true
	case sync:
		syncCmd.Parse(os.Args[2:])
		langMap = createLangsMap(syncLangs)
		cloneMissingRepos = true
	default:
		panic("unsupported subcommand " + os.Args[1])
	}

	cmd := cmd.New(l, ri, li, cloneMissingRepos, outputRepos, langMap)
	if err := cmd.Run(ctx); err != nil {
		panic(err)
	}
}

func createLangsMap(commaSeparated *string) *set.Set[string] {
	rv := set.New[string]()
	for _, lang := range strings.Split(*commaSeparated, ",") {
		if len(lang) > 0 {
			rv.Add(lang)
		}
	}
	return rv
}
