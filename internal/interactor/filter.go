package interactor

import (
	"fmt"
	"sgit/internal/set"
	"strings"
)

type Filter struct {
	langs, states *set.Set[string]
	names         []string
	forks         *bool
}

func NewFilter(langs, states, names string, forks *bool) (*Filter, error) {
	ss, err := statesSet(states)
	if err != nil {
		return nil, fmt.Errorf("failed to create state set: %w", err)
	}

	return &Filter{
		langs:  set.New(parseCommaSeparate(langs)...),
		states: ss,
		names:  parseCommaSeparate(names),
		forks:  forks,
	}, nil
}

func parseCommaSeparate(s string) []string {
	parts := strings.Split(s, ",")
	rv := make([]string, 0)
	for _, p := range parts {
		if len(p) > 0 {
			rv = append(rv, p)
		}
	}

	return rv
}

func (f Filter) ShouldInclude(rsp RepoStatePair) bool {
	if f.langs.Size() > 0 && !f.langs.Contains(rsp.Language) {
		return false
	}

	if f.forks != nil && !*f.forks && rsp.Fork {
		return false
	}

	if f.forks != nil && *f.forks && !rsp.Fork {
		return false
	}

	if f.states.Size() > 0 && !f.states.Contains(rsp.State.String()) {
		return false
	}

	if len(f.names) > 0 {
		for _, name := range f.names {
			if strings.Contains(rsp.Name, name) {
				return true
			}
		}

		return false
	}

	return true
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
