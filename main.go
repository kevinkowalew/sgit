package main

import (
	"sgit/internal/cmd"
	"log"
)

func main() {
	rc, err := cmd.NewRefreshCommand()
	if err != nil {
		log.Fatal(err.Error())
	}
	rc.Run()
}
