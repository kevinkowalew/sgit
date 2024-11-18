package cmd

import (
	"context"
	"fmt"
	"os"
	"sgit/internal/cmd/clone"
	"sgit/internal/cmd/create"
	"sgit/internal/cmd/ls"
	"sgit/internal/env"

	"github.com/spf13/cobra"
)

var (
	langs, states *string
	forks         *bool
)

var cmd = &cobra.Command{
	Use:   "sgit",
	Short: "git synchronization made simple",
	Long:  "git synchronization made simple",
}

func Execute() {
	env.AssertExists(env.GithubToken, env.GithubUsername, env.CodeHomeDir)

	ctx := context.Background()
	if err := cmd.ExecuteContext(ctx); err != nil {
		fmt.Println(err.Error())
		os.Exit(1)
	}
}

func init() {
	cmd.AddCommand(ls.Cmd)
	cmd.AddCommand(clone.Cmd)
	cmd.AddCommand(create.Cmd)
}
