package filesystem

import (
	"errors"
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

func (f Filesystem) DeleteDir(path string) error {
	if path == "" {
		return errors.New("path is empty, skipping for safety")
	}

	exists, err := f.Exists(path)
	if err != nil {
		return fmt.Errorf("filesystem.Exists failed: %w", err)
	}

	if !exists {
		return nil
	}

	cmd := "rm -r " + path
	_, err = execute(cmd, f.baseDir)
	return err
}

func (f Filesystem) MoveDir(existingPath, newPath string) error {
	cmd := fmt.Sprintf("mv %s %s", existingPath, newPath)
	_, err := execute(cmd, f.baseDir)
	return err
}

func execute(cmd, workingDir string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	b, err := c.Output()
	return string(b), err
}
