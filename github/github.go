package github

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"
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

func (g Github) GetPrimaryLanguageForRepo(ctx context.Context, owner, name string) (string, error) {
	e := fmt.Sprintf("/repos/%s/%s/languages", owner, name)
	langs, err := execute[map[string]int](ctx, http.MethodGet, e, g.token, nil)
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

func (g Github) GetAllRepos(ctx context.Context) ([]Repository, error) {
	// TODO: update to paginate correctly
	e := "/user/repos?affiliation=owner&per_page=100"
	repos, err := execute[[]Repository](ctx, http.MethodGet, e, g.token, nil)
	if err != nil {
		return nil, err
	}

	return *repos, nil
}

func (g Github) DeleteRepo(ctx context.Context, owner, name string) error {
	e := fmt.Sprintf("/repos/%s/%s", owner, name)
	_, err := execute[struct{}](ctx, http.MethodDelete, e, g.token, nil)
	return err
}

func (g Github) CreateRepo(ctx context.Context, name string, private bool) error {
	e := "/user/repos"
	json := struct {
		Name    string `json:"name"`
		Private bool   `json:"private"`
	}{name, private}

	_, err := execute[struct{}](ctx, http.MethodPost, e, g.token, json)
	return err
}

func execute[T any](ctx context.Context, verb, endpoint, token string, body any) (*T, error) {
	url := "https://api.github.com" + endpoint

	var r io.Reader
	if body != nil {
		b, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("json.Marshal failed: %w", err)
		}
		r = bytes.NewBuffer(b)
	}

	reqCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	req, err := http.NewRequestWithContext(reqCtx, verb, url, r)
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
	resBody, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	if res.StatusCode < 200 || res.StatusCode > 299 {
		msg := fmt.Sprintf("%s: %s", res.Status, body)
		return nil, errors.New(msg)
	}

	var t *T
	err = json.Unmarshal(resBody, &t)
	return t, err
}
