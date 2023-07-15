package filesystem

import (
	"fmt"
	"os"
	"sgit/internal/bash"
)

type FilesystemClient struct {
}

func NewClient() *FilesystemClient {
	return &FilesystemClient{}
}

func (fc FilesystemClient) CreateDirectory(path string) error {
	cmd := fmt.Sprintf("mkdir -p %s", path)
	return bash.Execute(cmd, path)
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
