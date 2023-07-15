package filesystem

import (
	"fmt"
	"internal/bash"
)

type FilesystemClient struct {
}

func NewFilesystem() *FilesystemClient {
	return &FilesystemClient{}
}

func (fc FilesystemClient) CreateDirectory(path string) error {
	cmd := fmt.Sprintf("mkdir -p %s", path)
	return bash.Execute(cmd, path)
}
