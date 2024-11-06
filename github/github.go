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

	Client struct {
		token, username string
	}
)

func NewClient(token, username string) *Client {
	return &Client{token, username}
}

func (c Client) GetPrimaryLanguageForRepo(ctx context.Context, name string) (string, error) {
	url := fmt.Sprintf("/repos/%s/%s/languages", c.username, name)
	langs, err := executeRequest[map[string]int](ctx, url, c.token)
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

func (c Client) GetAllRepos(ctx context.Context) ([]Repository, error) {
	// TODO: update to paginate correctly
	url := fmt.Sprintf("/user/repos?affiliation=owner&per_page=100")
	repos, err := executeRequest[[]Repository](ctx, url, c.token)
	if err != nil {
		return nil, err
	}

	rv := make([]Repository, 0)
	for _, repo := range *repos {
		// TODO: add flag to skip forks
		rv = append(rv, repo)
	}
	return rv, err
}

func (c Client) GetCommitHash(name, branch string) (string, error) {
	// TODO: implement me
	return "", nil
}

func executeRequest[T any](_ context.Context, endpoint, token string) (*T, error) {
	// TODO: actually use the context.Context instance in the request
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
