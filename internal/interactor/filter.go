package interactor

import (
	"fmt"
	"sgit/internal/set"
	"strings"
)

type Filter struct {
	langs  *set.Set[string]
	states *set.Set[State]
	names  []string
	forks  *bool
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

func (f Filter) Include(rsp RepoStatePair) bool {
	if f.langs.Size() > 0 && !f.langs.Contains(rsp.Language) {
		return false
	}

	if f.forks != nil && !*f.forks && rsp.Fork {
		return false
	}

	if f.forks != nil && *f.forks && !rsp.Fork {
		return false
	}

	if f.states.Size() > 0 && !f.states.Contains(rsp.State) {
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

func statesSet(commaSeparated string) (*set.Set[State], error) {
	normalize := func(s State) string {
		return strings.ToLower(s.String())
	}

	states := map[string]State{
		normalize(UpToDate):                         UpToDate,
		normalize(UncommittedChanges):               UncommittedChanges,
		normalize(NotGitRepo):                       NotGitRepo,
		normalize(NoRemoteRepo):                     NoRemoteRepo,
		normalize(IncorrectLanguageParentDirectory): IncorrectLanguageParentDirectory,
		normalize(NotCloned):                        NotCloned,
	}

	rv := set.New[State]()
	for _, s := range strings.Split(commaSeparated, ",") {
		if len(s) == 0 {
			continue
		}

		ls := strings.ToLower(s)
		found := false
		for normalizedName, state := range states {
			if strings.HasPrefix(normalizedName, ls) {
				found = true
				rv.Add(state)
			}
		}

		if !found {
			vs := make([]string, len(states))
			for _, s := range states {
				vs = append(vs, s.String())
			}

			return nil, fmt.Errorf(
				"\ninvalid -states flag: \"%s\" \nvalid flags: %s",
				s,
				strings.Join(vs, " "),
			)
		}

	}

	return rv, nil
}
