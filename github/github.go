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
	token string
	org   string
}

func NewClient(token, org string) *client {
	return &client{token, org}
}

func (c client) GetPrimaryLanguageForRepo(n string) (string, error) {
	url := fmt.Sprintf("/repos/%s/%s/languages", c.org, n)
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
	url := fmt.Sprintf("/orgs/%s/repos", c.org)
	repos, err := executeRequest[[]cmd.GithubRepository](url, c.token)
	if err != nil {
		return nil, err
	}
	return *repos, err
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
		msg := fmt.Sprintf("Non-200 status code returned (%s): %s", res.Status, body)
		return nil, errors.New(msg)
	}

	var t *T
	json.Unmarshal(body, &t)
	return t, nil
}
