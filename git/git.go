package git

import (
	"fmt"
	"os/exec"
	"sgit/internal/cmd"
	"strings"
)

type gitClient struct {
}

func NewClient() *gitClient {
	return &gitClient{}
}

func (c gitClient) CloneRepo(r cmd.GithubRepository, path string) error {
	cmd := fmt.Sprintf("git clone %s", r.SshUrl)
	_, err := execute(cmd, path)
	return err
}

func (c gitClient) HasUncommittedChanges(path string) (bool, error) {
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
	hasChanges, err := c.HasUncommittedChanges(path)
	if err != nil || !hasChanges {
		return err
	}

	_, err = execute("git add . && git commit -m 'work in progress' && git push", path)
	return err
}

func (c gitClient) StashLocalChanges(path string) error {
	_, err := execute("git add . && git stash", path)
	return err
}

func (c gitClient) ResetLocalChanges(path string) error {
	_, err := execute("git add . && git reset --hard", path)
	return err
}

func (c gitClient) PullLatest(path string) error {
	_, err := execute("git fetch && git pull", path)
	return err
}

func (c gitClient) HasMergeConflicts(path string) (bool, error) {
	return false, nil
}

func (c gitClient) GetCommitHashes(path string) ([]string, error) {
	// TODO: implement me
	//return execute("git log | head -n 1 | awk '{print($2)}'", path)
	return []string{}, nil
}

func (c gitClient) GetBranchName(path string) (string, error) {
	return "", nil
}

func execute(cmd, workingDir string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	b, err := c.Output()
	return string(b), err
}
