package github_client

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os/exec"
	"strings"
)

type GithubClient struct {
	token string
	org string
	baseDir string
}

func NewGithubClient(token, org string) (*GithubClient, error) {
	if token == "" {
		return nil, errors.New("Unable to create GithubClient: empty access token")
	}

	return &GithubClient{token,org, "~/code/"}, nil 
}

func (c GithubClient) CloneRepo(r Repository) {
	if r.Name() == "sgit" {
		return 
	}

	primaryLanaguage, err := c.GetPrimaryLanguageForRepo(r.Name())
	if err != nil {
		panic(err)
	}
	path := fmt.Sprintf("~/code/%s/%s/", strings.ToLower(primaryLanaguage), r.Name())
	cmd := fmt.Sprintf("git clone %s", r.CloneUrl)
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
	for k,v := range *langs {
		if v > maxLineCount {
			primaryLanguage = k
			maxLineCount = v
		}
	}
	return primaryLanguage, nil
}

func (c GithubClient) HasLocalChanges(path string) bool {
	gitCmd := "git status | grep 'nothing to commit, working tree clean'  | wc -l"

	cmd := exec.Command("bash", "-c", gitCmd)
	cmd.Dir = path
	o, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(o)) != "1"
}

func (c GithubClient) PushLocalChanges(path string) {
	if !c.HasLocalChanges(path) {
		return 
	}

	execCmd("git add . && git commit -m 'work in progress' && git push", path)
}

func (c GithubClient) StashLocalChanges(path string) {
	execCmd("git add . && git stash", path)
}

func (c GithubClient) ResetLocalChanges(path string) {
	execCmd("git add . && git reset --hard", path)
}

func (c GithubClient) PullLatestChanges(path string) {
	execCmd("git fetch && git pull", path)
}

func execCmd(cmd, workingDir string) {
	c := exec.Command("bash", "-c", cmd)
	c.Dir = workingDir

	_, err := c.Output()
	if err != nil {
		panic(err)
	}
}

type Repository struct {
	FullName string `json:"full_name"`
	CloneUrl string `json:"clone_url"`
}

func (r Repository) Name() string {
	p := strings.Split(r.FullName, "/")
	return p[len(p)-1]
}

func (c GithubClient) GetAllRepos() []Repository {
	req := c.createRequest(fmt.Sprintf("/orgs/%s/repos", c.org))
	repos := do[[]Repository](req)
	if repos == nil {
		return []Repository{}
	}
	return *repos
}

func (c GithubClient) createRequest(endpoint string) *http.Request {
	url := "https://api.github.com" + endpoint
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Authorization", "Bearer " + c.token)
	return req
}

func do[T any](req *http.Request) *T {
	client := &http.Client{}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err.Error())
	}

	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err.Error())
	}

	var t *T
	json.Unmarshal(body, &t)
	return t
}
