package git

import (
	"fmt"
	"os/exec"
	"sgit/internal/cmd"
	"strings"
)

type Git struct {
}

func New() *Git {
	return &Git{}
}

// TODO: add support for multiple remote repos
func (c Git) GetSshUrl(path string) (string, error) {
	gitCmd := "git remote | xargs git remote get-url"
	cmd := exec.Command("bash", "-c", gitCmd)
	cmd.Dir = path
	o, err := cmd.Output()

	if err != nil {
		return "", err
	}

	return strings.Split(string(o), "\n")[0], nil
}

func (c Git) Clone(r cmd.RemoteRepository, path string) error {
	cmd := fmt.Sprintf("git clone %s", r.SshUrl)
	_, err := execute(cmd, path)
	return err
}

func (c Git) HasUncommittedChanges(path string) (bool, error) {
	gitCmd := "git status | grep 'nothing to commit, working tree clean' | wc -l"

	cmd := exec.Command("bash", "-c", gitCmd)
	cmd.Dir = path
	o, err := cmd.Output()

	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(o)) != "1", nil
}

func (c Git) PushLocalChanges(path string) error {
	hasChanges, err := c.HasUncommittedChanges(path)
	if err != nil || !hasChanges {
		return err
	}

	_, err = execute("git add . && git commit -m 'work in progress' && git push", path)
	return err
}

func (c Git) PullLatest(path string) error {
	_, err := execute("git fetch && git pull", path)
	return err
}

func (c Git) HasMergeConflicts(path string) (bool, error) {
	return false, nil
}

func (c Git) GetCommitHashes(path string) ([]string, error) {
	// TODO: implement me
	//return execute("git log | head -n 1 | awk '{print($2)}'", path)
	return []string{}, nil
}

func (c Git) GetBranchName(path string) (string, error) {
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
