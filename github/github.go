package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"sgit/internal/cmd"
)

type client struct {
	token, username string
}

func NewClient(token, username string) *client {
	return &client{token, username}
}

func (c client) GetPrimaryLanguageForRepo(n string) (string, error) {
	url := fmt.Sprintf("/repos/%s/%s/languages", c.username, n)
	langs, err := executeRequest[map[string]int](url, c.token)
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

func (c client) GetAllRepos() ([]cmd.GithubRepository, error) {
	url := fmt.Sprintf("/user/repos?affiliation=owner&per_page=100")
	repos, err := executeRequest[[]cmd.GithubRepository](url, c.token)
	if err != nil {
		return nil, err
	}

	rv := make([]cmd.GithubRepository, 0)
	for _, repo := range *repos {
		// skip forks
		if repo.Fork {
			continue
		}
		rv = append(rv, repo)
	}
	return rv, err
}

func (c client) GetCommitHash(name, branch string) (string, error) {
	// TODO: implement me
	return "", nil
}

func executeRequest[T any](endpoint, token string) (*T, error) {
	url := "https://api.github.com" + endpoint
	req, err := http.NewRequest("GET", url, nil)
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
