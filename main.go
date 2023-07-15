package main

import (
	"sgit/internal/cmd"
)

func main() {
	rc, err := cmd.NewRefreshCommand()
	if err != nil {
		panic(err)
	}

	err = rc.Run()
	if err != nil {
		panic(err)
	}
}
