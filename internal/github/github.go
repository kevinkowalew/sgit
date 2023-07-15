package github

import (
	"encoding/json"
	"errors"
	"fmt"
	"internal/types"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type GithubClient struct {
	token     string
	org       string
	targetDir string
}

func NewGithubClient(token, org, targetDir string) (*GithubClient, error) {
	if token == "" {
		return nil, errors.New("Unable to create GithubClient: empty access token")
	}

	return &GithubClient{token, org, targetDir}, nil
}

func (c GithubClient) CloneRepo(r types.Repository) {
	primaryLanaguage, err := c.GetPrimaryLanguageForRepo(r.Name())
	if err != nil {
		panic(err)
	}

	path := fmt.Sprintf("%s/%s/%s", c.targetDir, strings.ToLower(primaryLanaguage), r.Name())
	cmd := fmt.Sprintf("mkdir -p %s", path)
	execCmd(cmd, "")

	cmd = fmt.Sprintf("git clone %s", r.SshUrl)
	execCmd(cmd, path)
}

func (c GithubClient) GetPrimaryLanguageForRepo(n string) (string, error) {
	url := fmt.Sprintf("/repos/%s/%s/languages", c.org, n)
	req := c.createRequest(url)
	langs := do[map[string]int](req)

	if langs == nil {
		return "", errors.New("Failed to fetch language usage metadata")
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

func execCmd(cmd, workingDir string) {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	_, err := c.Output()
	if err != nil {
		panic(err)
	}
}

func (c GithubClient) GetAllRepos() []types.Repository {
	req := c.createRequest(fmt.Sprintf("/orgs/%s/repos", c.org))
	repos := do[[]types.Repository](req)
	if repos == nil {
		return []types.Repository{}
	}
	return *repos
}

func (c GithubClient) createRequest(endpoint string) *http.Request {
	url := "https://api.github.com" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer "+c.token)
	return req
}

func do[T any](req *http.Request) *T {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		log.Fatalf("Failed to fetch repos from github (%s)", res.Status)
	}

	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	var t *T
	json.Unmarshal(body, &t)
	return t
}
