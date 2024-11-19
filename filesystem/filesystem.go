package filesystem

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
)

type Filesystem struct {
	baseDir string
}

func New(baseDir string) *Filesystem {
	return &Filesystem{baseDir}
}

func (f Filesystem) CreateDirectory(relativePath string) error {
	cmd := fmt.Sprintf("mkdir -p %s", relativePath)
	_, err := execute(cmd, "")
	return err
}

func (f Filesystem) Exists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil {
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}

func (f Filesystem) ListDirectories() ([]string, error) {
	cmd := "ls -l -d */*/*/ | awk '{print($9)}'"
	output, err := execute(cmd, f.baseDir)
	if err != nil {
		return nil, err
	}

	dirs := make([]string, 0)
	for _, line := range strings.Split(output, "\n") {
		if line != "" {
			dir := filepath.Join(f.baseDir, line)
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
