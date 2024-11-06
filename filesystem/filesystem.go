package filesystem

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type FilesystemClient struct {
	baseDir string
}

func NewClient(baseDir string) *FilesystemClient {
	return &FilesystemClient{baseDir}
}

func (fc FilesystemClient) CreateDirectory(relativePath string) error {
	cmd := fmt.Sprintf("mkdir -p %s", relativePath)
	_, err := execute(cmd, "")
	return err
}

func (fc FilesystemClient) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (fc FilesystemClient) ListDirectories() ([]string, error) {
	cmd := "ls -l -d */*/ | awk '{print($9)}'"
	output, err := execute(cmd, fc.baseDir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			dir := filepath.Join(fc.baseDir, line)
			dirs = append(dirs, dir)
		}
	}

	return dirs, nil
}

func execute(cmd, workingDir string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	b, err := c.Output()
	return string(b), err
}
