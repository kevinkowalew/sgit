package main

import (
	"context"
	"flag"
	"fmt"
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

	lsCmd, lsLangs, lsStates, lsForks := createFlagSet("ls")
	syncCmd, syncLangs, syncStates, syncForks := createFlagSet("sync")

	if len(os.Args) < 2 {
		panic("You must specify a subcommand: ls, sync")
	}

	cloneMissingRepos, outputRepos := false, false
	var langsSet *set.Set[string]
	var statesSet *set.Set[string]
	var forks *bool

	switch os.Args[1] {
	case ls:
		lsCmd.Parse(os.Args[2:])
		langsSet = createLangSet(lsLangs)
		statesSet = createStatesSet(lsStates)
		forks = createForks(*lsForks)
		outputRepos = true
	case sync:
		syncCmd.Parse(os.Args[2:])
		langsSet = createLangSet(syncLangs)
		statesSet = createStatesSet(syncStates)
		forks = createForks(*syncForks)
		cloneMissingRepos = true
	default:
		panic("unsupported subcommand " + os.Args[1])
	}

	cmd := cmd.New(l, ri, li, cloneMissingRepos, outputRepos, langsSet, statesSet, forks)
	if err := cmd.Run(ctx); err != nil {
		panic(err)
	}
}

func createFlagSet(cmd string) (*flag.FlagSet, *string, *string, *string) {
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	langs := fs.String("langs", "", "comma-separated list of languages to target")
	states := fs.String("states", "", "comma-separated list of states to target")
	fork := fs.String("forks", "", "bool flag to conditionally target fork vs. non-forked repos")
	return fs, langs, states, fork
}

func createLangSet(commaSeparated *string) *set.Set[string] {
	rv := set.New[string]()
	for _, lang := range strings.Split(*commaSeparated, ",") {
		if len(lang) > 0 {
			rv.Add(lang)
		}
	}
	return rv
}

func createStatesSet(commaSeparated *string) *set.Set[string] {
	states := []string{
		cmd.UpToDate.String(),
		cmd.UncommittedChanges.String(),
		cmd.NotGitRepo.String(),
		cmd.NoRemoteRepo.String(),
		cmd.IncorrectLanguageParentDirectory.String(),
		cmd.NotCloned.String(),
	}

	rv := set.New[string]()
	for _, s := range strings.Split(*commaSeparated, ",") {
		if len(s) == 0 {
			continue
		}

		found := false
		for _, state := range states {
			if strings.Contains(state, s) {
				found = true
				rv.Add(state)
			}
		}

		if !found {
			msg := fmt.Sprintf(
				"\ninvalid -states flag: \"%s\" \nvalid flags: %s",
				s,
				strings.Join(states, " "),
			)
			panic(msg)
		}

	}

	return rv
}

func createForks(val string) *bool {
	if len(val) == 0 {
		return nil
	}

	var rv bool
	switch val {
	case "true":
		rv = true
		return &rv
	case "false":
		rv = false
		return &rv
	default:
		msg := fmt.Sprintf(
			"invalid -forks flag: \"%s\" \nvalid flags: %s",
			val,
			"true false",
		)
		panic(msg)
	}
}
