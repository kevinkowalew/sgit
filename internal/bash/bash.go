package bash

import (
	"os/exec"
)

func Execute(cmd, workingDir string) error {
	c := exec.Command("bash", "-c", cmd)
	if workingDir != "" {
		c.Dir = workingDir
	}

	_, err := c.Output()
	return err
}
