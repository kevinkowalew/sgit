package git

import (
	"os/exec"
	"strings"
)

type GitClient struct {
}

func NewGitClient() *GitClient {
	return &GitClient{}
}

func (c GitClient) HasLocalChanges(path string) (bool, error) {
	gitCmd := "git status | grep 'nothing to commit, working tree clean'  | wc -l"

	cmd := exec.Command("bash", "-c", gitCmd)
	cmd.Dir = path
	o, err := cmd.Output()

	if err != nil {
		panic(err)
	}

	return strings.TrimSpace(string(o)) != "1", nil
}

func (c GitClient) PushLocalChanges(path string) error {
	hasChanges, err := c.HasLocalChanges(path)
	if err != nil {
		return err
	}
	if !hasChanges {
		return nil
	}

	return execCmd("git add . && git commit -m 'work in progress' && git push", path)
}

func (c GitClient) StashLocalChanges(path string) error {
	return execCmd("git add . && git stash", path)
}

func (c GitClient) ResetLocalChanges(path string) error {
	return execCmd("git add . && git reset --hard", path)
}

func (c GitClient) PullLatestChanges(path string) error {
	return execCmd("git fetch && git pull", path)
}

func execCmd(cmd, workingDir string) error {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	_, err := c.Output()
	return err
}
