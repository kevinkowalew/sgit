package git

import (
	"fmt"
	"os/exec"
	"sgit/internal/bash"
	"sgit/internal/types"
	"strings"
)

type gitClient struct {
}

func NewClient() *gitClient {
	return &gitClient{}
}

func (c gitClient) CloneRepo(r types.GithubRepository, path string) error {
	cmd := fmt.Sprintf("git clone %s", r.SshUrl)
	return bash.Execute(cmd, path)
}

func (c gitClient) HasLocalChanges(path string) (bool, error) {
	gitCmd := "git status | grep 'nothing to commit, working tree clean'  | wc -l"

	cmd := exec.Command("bash", "-c", gitCmd)
	cmd.Dir = path
	o, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(o)) != "1", nil
}

func (c gitClient) PushLocalChanges(path string) error {
	hasChanges, err := c.HasLocalChanges(path)
	if err != nil || !hasChanges {
		return err
	}

	return bash.Execute("git add . && git commit -m 'work in progress' && git push", path)
}

func (c gitClient) StashLocalChanges(path string) error {
	return bash.Execute("git add . && git stash", path)
}

func (c gitClient) ResetLocalChanges(path string) error {
	return bash.Execute("git add . && git reset --hard", path)
}

func (c gitClient) PullLatestChanges(path string) error {
	return bash.Execute("git fetch && git pull", path)
}
