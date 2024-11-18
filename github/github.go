package github

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
)

type (
	Repository struct {
		FullName string `json:"full_name"`
		SshUrl   string `json:"ssh_url"`
		Fork     bool   `json:"fork"`
		Owner    *Owner `json:"owner"`
	}

	Owner struct {
		Login string `json:"login"`
	}

	Github struct {
		token, username string
	}
)

func New(token, username string) *Github {
	return &Github{token, username}
}

func (c Github) GetPrimaryLanguageForRepo(ctx context.Context, name string) (string, error) {
	e := fmt.Sprintf("/repos/%s/%s/languages", c.username, name)
	langs, err := executeRequest[map[string]int](ctx, http.MethodGet, e, c.token)
	if err != nil {
		return "", err
	}

	primaryLanguage := "unknown"
	maxLineCount := -1
	for k, v := range *langs {
		if v > maxLineCount {
			primaryLanguage = k
			maxLineCount = v
		}
	}
	return primaryLanguage, nil
}

func (c Github) GetAllRepos(ctx context.Context) ([]Repository, error) {
	// TODO: update to paginate correctly
	e := fmt.Sprintf("/user/repos?affiliation=owner&per_page=100")
	repos, err := executeRequest[[]Repository](ctx, http.MethodGet, e, c.token)
	if err != nil {
		return nil, err
	}

	rv := make([]Repository, 0)
	for _, repo := range *repos {
		rv = append(rv, repo)
	}
	return rv, err
}

func (c Github) RepoExists(ctx context.Context) ([]Repository, error) {
	return nil, nil
}

func executeRequest[T any](_ context.Context, verb, endpoint, token string) (*T, error) {
	// TODO: update context to include timeout no request
	url := "https://api.github.com" + endpoint
	req, err := http.NewRequest(verb, url, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Add("Authorization", "Bearer "+token)

	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		return nil, err
	}

	defer res.Body.Close()
	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != 200 {
		msg := fmt.Sprintf("%s: %s", res.Status, body)
		return nil, errors.New(msg)
	}

	var t *T
	json.Unmarshal(body, &t)
	return t, nil
}
