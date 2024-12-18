package cmd

import (
	"context"
	"fmt"
	"os"
	"sgit/internal/cmd/clone"
	"sgit/internal/cmd/create"
	del "sgit/internal/cmd/delete"
	"sgit/internal/cmd/ls"

	"github.com/spf13/cobra"
)

var cmd = &cobra.Command{
	Use:   "sgit",
	Short: "git made simple",
	Long:  "git made simple",
}

func Execute() {
	assert("GITHUB_TOKEN")
	assert("GITHUB_USERNAME")
	assert("CODE_HOME_DIR")

	ctx := context.Background()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func assert(envVar string) {
	if _, ok := os.LookupEnv(envVar); !ok {
		panic("Unset environment variable: " + envVar)
	}
}

func init() {
	cmd.AddCommand(ls.Cmd)
	cmd.AddCommand(clone.Cmd)
	cmd.AddCommand(create.Cmd)
	cmd.AddCommand(del.Cmd)
}
