package interactor

import (
	"fmt"
	"sgit/internal/set"
	"strings"
)

type Filter struct {
	langs, states *set.Set[string]
	forks         *bool
}

func NewFilter(langs, states string, forks *bool) (*Filter, error) {
	langsSet := set.New[string]()
	langsP := strings.Split(langs, ",")
	for _, lang := range langsP {
		if len(lang) > 0 {
			langsSet.Add(lang)
		}
	}

	ss, err := statesSet(states)
	if err != nil {
		return nil, fmt.Errorf("failed to create state set: %w", err)
	}

	return &Filter{langsSet, ss, forks}, nil
}

func (f Filter) ShouldIncludeRepoStatePair(rsp RepoStatePair) bool {
	if !f.ShouldIncludeRepo(rsp.Repo) {
		return false
	}

	return f.states.Size() == 0 || f.states.Contains(rsp.State.String())
}

func (f Filter) ShouldIncludeRepo(r Repo) bool {
	if f.langs.Size() > 0 && !f.langs.Contains(r.Language) {
		return false
	}

	if f.forks != nil && !*f.forks && r.Fork {
		return false
	} else if f.forks != nil && *f.forks && !r.Fork {
		return false
	} else {
		return true
	}
}

func statesSet(commaSeparated string) (*set.Set[string], error) {
	states := []string{
		UpToDate.String(),
		UncommittedChanges.String(),
		NotGitRepo.String(),
		NoRemoteRepo.String(),
		IncorrectLanguageParentDirectory.String(),
		NotCloned.String(),
	}

	rv := set.New[string]()
	for _, s := range strings.Split(commaSeparated, ",") {
		if len(s) == 0 {
			continue
		}

		found := false
		for _, state := range states {
			if strings.HasPrefix(state, s) {
				found = true
				rv.Add(state)
			}
		}

		if !found {
			return nil, fmt.Errorf(
				"\ninvalid -states flag: \"%s\" \nvalid flags: %s",
				s,
				strings.Join(states, " "),
			)
		}

	}

	return rv, nil
}
