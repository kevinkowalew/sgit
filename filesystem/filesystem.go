package filesystem

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type FilesystemClient struct {
}

func NewClient() *FilesystemClient {
	return &FilesystemClient{}
}

func (fc FilesystemClient) CreateDirectory(path string) error {
	cmd := fmt.Sprintf("mkdir -p %s", path)
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

func (fc FilesystemClient) ListDirectories(path string, depth int) ([]string, error) {
	cmd := fmt.Sprintf("ls -l -d */*/ %s | awk '{print($9)}'", path)
	output, err := execute(cmd, "")
	if err != nil {
		return nil, err
	}

	return strings.Split(output, "\n"), nil
}

func execute(cmd, workingDir string) (string, error) {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	b, err := c.Output()
	return string(b), err
}
